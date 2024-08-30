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

type Gitea struct {
	*gitea.Client

	RepoName  string
	RepoOwner string
}

func (g *Gitea) UploadBugs(bc *BugzClient) {

	stmt, err := prepareStmt(bc)
	util.CheckFatal("failed to prepare statement", err)
	defer stmt.Finalize()

	err = processBugRows(stmt, func(bug Bug, rawJSON string) error {
		return g.createGiteaIssue(bug, rawJSON)
	})
	util.CheckFatal("error processing bug rows", err)
}

func New() Gitea {
	apiToken := os.Getenv("GITEA_TOKEN")
	if apiToken == "" {
		util.Fatal("GITEA_TOKEN environment variable is not set")
	}

	client, err := gitea.NewClient("https://gitcvt.hugpoint.tech", gitea.SetToken(apiToken))
	util.CheckFatal("error creating Gitea client", err)

	repoOwner, repoName, err := getRepoDetails()
	util.CheckFatal("failed to get repository details", err)

	return Gitea{
		Client:    client,
		RepoName:  repoName,
		RepoOwner: repoOwner,
	}
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

func (g *Gitea) createGiteaIssue(bug Bug, rawJSON string) error {
	issueBody := fmt.Sprintf(
		"ID: %d\n\nCreation Time: %s\n\nCreator: %s\n\nDetails:\n%s\n",
		bug.ID, bug.CreationTime, bug.Creator, rawJSON,
	)

	issue, _, err := g.CreateIssue(g.RepoOwner, g.RepoName, gitea.CreateIssueOption{
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
