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
		fmt.Sprintf(os.Getenv("JQL_QUERY"), os.Getenv("JIRA_PROJECT_ID"), fixVersion, strings.ReplaceAll(fixVersion, " ", "_")))
}

func checkMissingCommits(fxVersion string) (*[]string) {
	var missingCommitsList []string

	mainBranchName := os.Getenv("MAIN_BRANCH")	

	makeSureBranchFxIsThereCommand := fmt.Sprintf("git -C %s fetch origin %s:%s", os.Getenv("REPO_PATH"), fxVersion, fxVersion)
	makeSureMainBranchIsThereCommand := fmt.Sprintf("git -C %s fetch origin %s:%s", os.Getenv("REPO_PATH"), mainBranchName, mainBranchName)
	command := fmt.Sprintf("git -C %s cherry -v %s %s", os.Getenv("REPO_PATH"), fxVersion, mainBranchName)

	runcmdIgnoreErrors(makeSureBranchFxIsThereCommand, true)
	runcmdIgnoreErrors(makeSureMainBranchIsThereCommand, true)
	result := runcmd(command, true)

	resultString := bytes.NewBuffer(result).String()

	scanner := bufio.NewScanner(strings.NewReader(resultString))
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		if strings.HasPrefix(fields[0], "+") && strings.HasPrefix(fields[2], "SOL-") {
			taskId := strings.Split(fields[2], "_")[0]
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

func main() {
	godotenv.Load(".env")
	
	jiraTagName := os.Getenv("JIRA_TAG")
	gitBranchName := os.Getenv("GIT_BRANCH_NAME")
	completedTasksChannel := make(chan *[]string, 1)

	// Make sure we trust the repo we wanna check
	runcmd(fmt.Sprintf("git config --global --add safe.directory %s", os.Getenv("REPO_PATH")), true)

	fmt.Printf("Lacze z jira...\n")
	client := connectToJira()

	go searchForTasks(client, jiraTagName, completedTasksChannel)
	missingCommits := checkMissingCommits(gitBranchName)

	fmt.Printf("Pobieram dane z jiry...\n\n")
	completedTasks := <-completedTasksChannel

	fmt.Printf("Ukonczone zadania z Jiry dla tagu %s: \n%s\n\n", jiraTagName, *completedTasks)
	fmt.Printf("Commity ktorych nie ma na %s: \n%s\n\n", gitBranchName, *missingCommits)

	testSet := make(map[string]bool)
	for _, e := range *missingCommits {

		if _, prs := testSet[e]; prs {
			panic("Zduplikowane ID w brakujacych commitach - CO SIE DZIEJE?!?")		
		}

		testSet[e] = true
	}

	noIssues := true
	fmt.Printf("Wykrywam brakujace taski dla wersji %s...\n", gitBranchName)

	for _, e := range *completedTasks {
		if _, prs := testSet[e]; prs {
			fmt.Printf(" ! %s \n", e)
			noIssues = false
		}
	}

	if noIssues {
		fmt.Printf(" Wszystko git <3\n")
	} else {
		panic(" Prosze pomergowac brakujace commity </3")
	}
}

func runcmdIgnoreErrors(cmd string, shell bool) {
	fmt.Printf("Wolam komende (i ignoruje bledy): %s\n", cmd)

    if shell {
        _, err := exec.Command("bash", "-c", cmd).Output()

        if err != nil {
        	fmt.Printf(" - Blad (ignorowany): %s\n", err)
        }
        return
    }

    _, err := exec.Command(cmd).Output()
    if err != nil {
    	fmt.Printf(" - Blad (ignorowany): %s\n", err)
    }

    return
}

func runcmd(cmd string, shell bool) []byte {
	fmt.Printf("Wolam komende: %s\n", cmd)

    if shell {
        out, err := exec.Command("bash", "-c", cmd).Output()
        if err != nil {
            log.Fatal(err)
            panic("some error found")
        }
        return out
    }

    out, err := exec.Command(cmd).Output()
    if err != nil {
        log.Fatal(err)
    }

    return out
}