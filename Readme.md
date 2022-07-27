# Uruchomienie skryptu

```go run main.go```

# Jira Token
Aby wygenerowac swoj token wejdz na strone:
https://id.atlassian.com/manage-profile/security/api-tokens
Ten wygenerowany token przypisz do JIRA_TOKEN
do JIRA_USER adres email konta

# Konfiguracja
Plik .env o nastepujacym formacie:

```
###
### Program Settings - Input to the program
###

JIRA_TAG=""
GIT_BRANCH_NAME=""


###
### User Settings - Please fill out!
###

JIRA_USER="adres@email.com"
JIRA_TOKEN="token_jiry"
REPO_PATH="absolutna_lokalizacja_repo"

###
### Repository Settings - Should be no need to change
###
JIRA_URL="adres_url"
MAIN_BRANCH="main_branch_name"
JIRA_PROJECT_ID="ID projektu na Jirze"
JQL_QUERY="project = '%s' AND (fixVersion = '%s' OR labels = '%s') and (Status = 'RESOLVED' OR Status = 'CLOSED') and component = 'Android'"
```