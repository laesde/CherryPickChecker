package main

import (
	"bufio"
	"bytes"
	"strings"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/joho/godotenv"
	jira "gopkg.in/andygrunwald/go-jira.v1"
)

func connectToJira() (*jira.Client) {
	tp := jira.BasicAuthTransport {
		Username: os.Getenv("JIRA_USER"),
		Password: os.Getenv("JIRA_TOKEN"),
	}

	client, err := jira.NewClient(tp.Client(), os.Getenv("JIRA_URL"))

	if  err != nil {
		panic(err)
	} 

	return client
}

func getProjects(client *jira.Client) {
	// use the Project domain to list all projects
	projectList, _, err := client.Project.GetList()
	if err != nil {
		log.Fatal(err)
	}

	// Range over the projects and print the key and name
	for _, project := range *projectList {
		fmt.Printf("%s: %s\n", project.Key, project.Name)
	}
}

func runJQLQuery(client *jira.Client, jquQuery string) (*[]string) {
	issues, _, err := client.Issue.Search(jquQuery, nil)
	if err != nil {
		panic(err)
	}

	var result []string

	for _, i := range issues {
		result = append(result, i.Key)
	}

	return &result
}

func searchForTasks(client *jira.Client, fixVersion string, result chan *[]string) {
	result <- runJQLQuery(
		client, 
		fmt.Sprintf(os.Getenv("JQL_QUERY"), os.Getenv("JIRA_PROJECT_ID"), fixVersion))
}

// returns set of ticket that headBranch has but upstreamBranch doesn't
func diffJiraTicketsBetween(upstreamBranch string, headBranch string) (*[]string) {
	var missingCommitsList []string

	command := fmt.Sprintf("git -C %s log  --format=%q %s..%s", os.Getenv("REPO_PATH"), "%s", upstreamBranch, headBranch)
	result := runcmd(command, true)

	resultString := bytes.NewBuffer(result).String()
	scanner := bufio.NewScanner(strings.NewReader(resultString))
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		if strings.HasPrefix(fields[0], "SOL-") {
			taskId := strings.Split(fields[0], "_")[0]
			explodedTaskId := strings.Split(taskId, "-")

			missingCommitsList = append(
				missingCommitsList, 
				fmt.Sprintf("%s-%s", explodedTaskId[0], explodedTaskId[1]))
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
	return &missingCommitsList
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func prepareSources(sources [2]string) {
	for _, source := range sources {
		makeSureMainBranchIsThereCommand := fmt.Sprintf("git -C %s fetch origin %s", os.Getenv("REPO_PATH"), source)
		runcmdIgnoreErrors(makeSureMainBranchIsThereCommand, true)
	}
}

func checkMissingCommits(fxVersion string, mainBranchName string) (*[]string) {
	missingCommitOnFxBranch := diffJiraTicketsBetween(fxVersion, mainBranchName)
	missingCommitsOnMainBranch := diffJiraTicketsBetween(mainBranchName, fxVersion)
	result := []string{}
    for _, v := range *missingCommitOnFxBranch {
		// remove false positives
		if contains(*missingCommitsOnMainBranch, v) == false {
			result = append(result, v)	
		}
    }
	return &result
}


func main() {
	godotenv.Load(".env")
	
	jiraTagName := os.Getenv("JIRA_TAG")
	gitBranchName := os.Getenv("GIT_BRANCH_NAME")
	gitMainBranchName := os.Getenv("MAIN_BRANCH")
	completedTasksChannel := make(chan *[]string, 1)

	// fmt.Printf("repo path: %s\n", os.Getenv("REPO_PATH"))
		// Make sure we trust the repo we wanna check
	runcmd(fmt.Sprintf("git config --global --add safe.directory %s", os.Getenv("REPO_PATH")), true)
	branches := [...]string{gitBranchName, gitMainBranchName}
	prepareSources(branches)

	//fmt.Printf("Connecting to Jira...\n")
	client := connectToJira()

	go searchForTasks(client, jiraTagName, completedTasksChannel)
	
	missingCommits := checkMissingCommits(gitBranchName, gitMainBranchName)

	//fmt.Printf("Getting data from Jira...\n\n")
	completedTasks := <-completedTasksChannel

	//fmt.Printf("Finished tasks from Jira for tag %s: \n%s\n\n", jiraTagName, *completedTasks)
	//fmt.Printf("Commits missing from %s: \n%s\n\n", gitBranchName, *missingCommits)

	testSet := make(map[string]bool)
	for _, e := range *missingCommits {

		if _, prs := testSet[e]; prs {
			// panic("Duplicated ID in missing commits - This should not happen")		
		}

		testSet[e] = true
	}

	noIssues := true
	fmt.Printf("Detecting missing tasks for version %s...\n", gitBranchName)

	for _, e := range *completedTasks {
		if _, prs := testSet[e]; prs {
			fmt.Printf(" ! %s \n", e)
			noIssues = false
		}
	}

	if noIssues {
		fmt.Printf(" Everything fine <3\n")
	} else {
		fmt.Printf("  Please cherry-pick missing commits </3\n")
		os.Exit(1)
	}
}

func runcmdIgnoreErrors(cmd string, shell bool) {
	//fmt.Printf("Calling cmd (and ignoring exceptions): %s\n", cmd)

    if shell {
        _, err := exec.Command("bash", "-c", cmd).Output()

        if err != nil {
        	//fmt.Printf(" - Exception (ignored): %s\n", err)
        }
        return
    }

    _, err := exec.Command(cmd).Output()
    if err != nil {
    	//fmt.Printf(" - Exception (ignored): %s\n", err)
    }

    return
}

func runcmd(cmd string, shell bool) []byte {
	//fmt.Printf("Calling cmd: %s\n", cmd)

    if shell {
        out, err := exec.Command("bash", "-c", cmd).Output()
        if err != nil {
            //log.Fatal(err)
            panic("some error found")
        }
        return out
    }

    out, err := exec.Command(cmd).Output()
    if err != nil {
        //log.Fatal(err)
    }

    return out
}