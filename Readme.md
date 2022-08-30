# How to integrate the CherryPickChecker
1. Clone repo
2. Prepare .env file 
3. Run tool by command listed below:

```
go isntall
go mod tidy
export RESULT_TEXT=`go run main.go`
```

RESULT_TEXT will contain the text output that can be sent via SLACK API
Example value:
```
Detecting missing tasks for version tve_9.5...
! SOL-33894
! SOL-33858
! SOL-33478
 Please cherry-pick missing commits </3
```


# Jira Token
To generate jira token check the website below:
https://id.atlassian.com/manage-profile/security/api-tokens

JIRA_TOKEN - jira token
JIRA_USER - email of account that created the token

# Configuration
Put all configuration into the .env file

```
###
### Program Settings - Input to the program
###

JIRA_TAG=""
GIT_BRANCH_NAME=""


###
### User Settings - Please fill out!
###

JIRA_USER="jira_email"
JIRA_TOKEN="jira_token"
REPO_PATH="absolute_repo_path"

###
### Repository Settings - Should be no need to change
###
JIRA_URL="jira_url"
MAIN_BRANCH="main_branch_name"
JIRA_PROJECT_ID="project_id_from_jira"
JQL_QUERY="project = '%s' AND (fixVersion = '%s') and (Status = 'RESOLVED' OR Status = 'CLOSED') and component = 'iOS'"
```
