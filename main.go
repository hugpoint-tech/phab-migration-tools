package main

import (
	"fmt"
	. "hugpoint.tech/freebsd/forge/bugz"
	giteacustom "hugpoint.tech/freebsd/forge/gitea"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("not enough arguments. use help")
		return
	}
	command := os.Args[1]

	bugzDatabasePath := "bugsNew.db"           // Specify the path to the database
	bc, err := NewBugzClient(bugzDatabasePath) // Create a BugzClient instance
	if err != nil {
		log.Fatalf("Failed to create BugzClient: %v", err)
	}

	switch command {
	case "bugzilla-download-bugs":
		err := bc.DownloadBugzillaBugs()
		if err != nil {
			log.Fatalf("Error downloading bugs: %v\n", err)
		}
	case "help":
		printHelp()
	case "bugzilla-list-bugs":
		err := bc.ListBugs()
		if err != nil {
			log.Fatalf("Error listing bugs: %v\n", err)
		}
	case "bugzilla-download-users":
		err := bc.DownloadBugzillaUsers()
		if err != nil {
			log.Fatalf("Error downloading users: %v\n", err)
		}
	case "bugzilla-show-bugs":
		if err := bc.ShowBugs(); err != nil {
			log.Fatalf("error showing bugs: %v\n", err)
		}
	case "bugzilla-download-comments":
		// Fetch bugs from the SQLite database
		bugs, err := FetchBugsFromDatabase(bc.Db)
		if err != nil {
			log.Fatalf("Error fetching bugs: %v", err)
		}

		// Convert bug IDs to a slice of int64 for processing
		var bugIDs []int64
		for _, bug := range bugs {
			bugIDs = append(bugIDs, int64(bug.ID))
		}

		// Download comments using the producer-consumer pattern
		err = bc.DownloadBugzillaComments(bugIDs)
		if err != nil {
			log.Fatalf("Error downloading comments: %v", err)
		}

		fmt.Println("Downloaded comments for all bugs successfully.")
	case "bugzilla-download-attachments":
		// Fetch bugs from the SQLite database
		bugs, err := FetchBugsFromDatabase(bc.Db)
		if err != nil {
			log.Fatalf("Error fetching bugs: %v", err)
		}

		// Convert bug IDs to a slice of int64 for processing
		var bugIDs []int64
		for _, bug := range bugs {
			bugIDs = append(bugIDs, int64(bug.ID))
		}

		// Download attachments using the producer-consumer pattern
		err = bc.DownloadBugzillaAttachments(bugIDs)
		if err != nil {
			log.Fatalf("Error downloading comments: %v", err)
		}

		fmt.Println("Downloaded attachments for all bugs successfully.")
	case "gitea-upload-bugs":

		var url string
		if len(os.Args) < 3 {
			url = giteacustom.DEFAULT_GITEA_URL
			fmt.Println("using default gitea url", url)
		} else {
			url = os.Args[2]
		}

		giteaClient := giteacustom.New(url)
		giteaClient.UploadBugs(bc)

	default:
		fmt.Println("invalid command")
		printHelp()
	}
}

func printHelp() { // not sure about functions descriptions
	fmt.Println("available commands:\n" +
		"bugzilla-download-bugs - downloads bugs from bugzilla\n" +
		"bugzilla-show-bugs - shows bugzilla bugs\n" +
		"bugzilla-list-bugs - displays downloaded bugs\n" +
		"bugzilla-download-users - downloads users from bugzilla\n" +
		"bugzilla-download-comments - downloads comments from bugs db\n" +
		"bugzilla-download-attachments - downloads attachments from bugs db\n" +
		"gitea-upload-bugs [url] - upload bugs to gitea from local bugzilla db\n" +
		"help - shows available commands")

}
