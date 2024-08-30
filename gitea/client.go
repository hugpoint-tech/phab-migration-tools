package gitea_custom

import (
	"code.gitea.io/sdk/gitea"
	"encoding/json"
	"fmt"
	. "hugpoint.tech/freebsd/forge/bugz"
	"hugpoint.tech/freebsd/forge/util"
	"os"
	"zombiezen.com/go/sqlite"
)

func UploadBugs(bc *BugzClient) error {

	repoOwner, repoName, err := getRepoDetails()
	if err != nil {
		return fmt.Errorf("failed to get repository details: %w", err)
	}

	stmt, err := prepareStmt(bc)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Finalize()

	// Initialize Gitea client
	client := NewGiteaClient()

	err = processBugRows(stmt, func(bug Bug, rawJSON string) error {
		return createGiteaIssue(client, repoOwner, repoName, bug, rawJSON)
	})
	if err != nil {
		return fmt.Errorf("error processing bug rows: %w", err)
	}

	return nil
}

func NewGiteaClient() *gitea.Client {
	apiToken := os.Getenv("GITEA_TOKEN")
	if apiToken == "" {
		util.Fatal("GITEA_TOKEN environment variable is not set")
	}

	client, err := gitea.NewClient("https://gitcvt.hugpoint.tech", gitea.SetToken(apiToken))
	util.CheckFatal("error creating Gitea client", err)

	return client
}

func getRepoDetails() (string, string, error) {
	repoOwner := os.Getenv("REPO_OWNER")
	repoName := os.Getenv("REPO_NAME")

	if repoOwner == "" || repoName == "" {
		return "", "", fmt.Errorf("REPO_OWNER and REPO_NAME environment variables must be set")
	}
	return repoOwner, repoName, nil
}

func prepareStmt(bc *BugzClient) (*sqlite.Stmt, error) {
	stmt, err := bc.Db.Prepare("SELECT OtherFieldsJSON FROM bugs")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	return stmt, nil
}

func createGiteaIssue(client *gitea.Client, repoOwner, repoName string, bug Bug, rawJSON string) error {
	issueBody := fmt.Sprintf(
		"ID: %d\n\nCreation Time: %s\n\nCreator: %s\n\nDetails:\n%s\n",
		bug.ID, bug.CreationTime, bug.Creator, rawJSON,
	)

	issue, _, err := client.CreateIssue(repoOwner, repoName, gitea.CreateIssueOption{
		Title: bug.Summary,
		Body:  issueBody,
	})
	if err != nil {
		return fmt.Errorf("failed to create issue: %w", err)
	}
	fmt.Printf("Issue created: %s\n", issue.URL)
	return nil
}

func processBugRows(stmt *sqlite.Stmt, processFunc func(bug Bug, rawJSON string) error) error {
	for {
		hasRow, err := stmt.Step()
		if err != nil {
			return fmt.Errorf("failed to step through rows: %w", err)
		}
		if !hasRow {
			break
		}

		OtherFieldsJSON := stmt.ColumnText(0)

		var bug Bug
		if err := json.Unmarshal([]byte(OtherFieldsJSON), &bug); err != nil {
			fmt.Printf("Failed to parse JSON for row %d: %v\nData: %s", stmt.ColumnInt(0), err, OtherFieldsJSON)
			continue
		}

		if err := processFunc(bug, OtherFieldsJSON); err != nil {
			fmt.Printf("Failed to process bug ID %d: %v\nData: %s", bug.ID, err, OtherFieldsJSON)
			continue
		}
	}
	return nil
}
