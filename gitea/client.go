package gitea_custom

import (
	"code.gitea.io/sdk/gitea"
	"encoding/json"
	"fmt"
	. "hugpoint.tech/freebsd/forge/bugz"
	"log"
	"os"
)

func GiteaGetBugz(bc *BugzClient) error {

	// Set environment variables for repo
	repoOwner := os.Getenv("REPO_OWNER")
	repoName := os.Getenv("REPO_NAME")

	if repoOwner == "" || repoName == "" {
		fmt.Println("REPO_OWNER and REPO_NAME environment variables must be set")
		return nil
	}

	stmt, err := bc.Db.Prepare("SELECT OtherFieldsJSON FROM bugs")
	if err != nil {
		log.Fatalf("Failed to prepare statement: %v", err)
	}
	defer stmt.Finalize()

	// Initialize Gitea client
	client, err := NewGiteaClient()
	if err != nil {
		log.Fatalf("Failed to create Gitea client: %v", err)
	}

	for {
		hasRow, err := stmt.Step()
		if err != nil {
			log.Fatalf("Failed to step through rows: %v", err)
		}
		if !hasRow {
			break
		}

		OtherFieldsJSON := stmt.ColumnText(0)

		var bug Bug
		if err := json.Unmarshal([]byte(OtherFieldsJSON), &bug); err != nil {
			log.Printf("Failed to parse JSON: %v", err)
			continue
		}

		issue, _, err := client.CreateIssue(repoOwner, repoName, gitea.CreateIssueOption{
			Title: bug.Summary,
			Body:  fmt.Sprintf("Details:\n%s", OtherFieldsJSON),
		})
		if err != nil {
			log.Printf("Failed to create issue: %v\n", err)
		} else {
			fmt.Printf("Issue created: %s\n", issue.URL)
		}
	}

	return nil
}

func NewGiteaClient() (*gitea.Client, error) {
	apiToken := os.Getenv("GITEA_TOKEN")
	if apiToken == "" {
		fmt.Errorf("GITEA_TOKEN environment variable is not set")
	}

	client, err := gitea.NewClient("https://gitcvt.hugpoint.tech", gitea.SetToken(apiToken))
	if err != nil {
		fmt.Errorf("Error creating Gitea client: %v", err)
	}

	return client, nil
}
