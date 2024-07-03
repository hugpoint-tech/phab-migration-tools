package main

import (
	"fmt"
	. "hugpoint.tech/freebsd/forge/bugz"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("not enough arguments. use help")
		return
	}
	command := os.Args[1]

	databasePath := "bugsNew.db"           // Specify the path to the database
	bc, err := NewBugzClient(databasePath) // Create a BugzClient instance
	if err != nil {
		log.Fatalf("Failed to create BugzClient: %v", err)
	}

	switch command {
	case "bugzilla-download-bugs":
		err := bc.DownloadBugzillaBugs()
		if err != nil {
			fmt.Printf("Error downloading bugs: %v\n", err)
		}
	case "help":
		printHelp()
	case "bugzilla-list-bugs":
		err := bc.ListBugs()
		if err != nil {
			fmt.Printf("Error listing bugs: %v\n", err)
		}
	case "bugzilla-download-users":
		err := bc.DownloadBugzillaUsers()
		if err != nil {
			fmt.Printf("Error downloading users: %v\n", err)
		}
	case "bugzilla-show-bugs":
		if err := bc.ShowBugs(); err != nil {
			fmt.Printf("error showing bugs: %v\n", err)
		}
	case "bugzilla-download-comments":
		// Fetch bugs from the SQLite database
		bugs, err := FetchBugsFromDatabase(bc.Db)
		if err != nil {
			log.Fatalf("Error fetching bugs: %v", err)
		}
		// Download comments for each bug
		for _, bug := range bugs {
			err := bc.DownloadBugzillaComments(int64(bug.ID))
			if err != nil {
				log.Printf("Error downloading comments for bug %d: %v", bug.ID, err)
			}
		}
		fmt.Println("Downloaded comments for all bugs successfully.")
	case "bugzilla-download-attachments":
		// Fetch bugs from the SQLite database
		bugs, err := FetchBugsFromDatabase(bc.Db)
		if err != nil {
			log.Fatalf("Error fetching bugs: %v", err)
		}
		// Download attachments for each bug
		for _, bug := range bugs {
			err := bc.DownloadBugzillaAttachments(int64(bug.ID))
			if err != nil {
				log.Printf("Error downloading attachments for bug %d: %v", bug.ID, err)
			}
		}
		fmt.Println("Downloaded attachments for all bugs successfully.")
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
		"help - shows available commands")

}
