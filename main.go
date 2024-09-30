package main

import (
	"fmt"
	. "hugpoint.tech/freebsd/forge/bugzilla"
	"hugpoint.tech/freebsd/forge/database"
	"hugpoint.tech/freebsd/forge/forgejo"
	"hugpoint.tech/freebsd/forge/operations"
	"os"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("not enough arguments. use help")
		return
	}
	command := os.Args[1]
	db := database.New("migrator.db")
	bugzClient := NewClient()

	switch command {
	case "bugzilla-download-bugs":
		startTime := time.Now() // Record the start time

		// Call the download function
		operations.DownloadBugzillaBugs(&bugzClient, db)

		// Calculate and print the time taken
		duration := time.Since(startTime)
		fmt.Printf("DownloadBugzillaBugs completed in %s\n", duration)
	case "help":
		printHelp()

	case "bugzilla-download-users":
		//err := bugzClient.DownloadBugzillaUsers()
		//if err != nil {
		//	log.Fatalf("Error downloading users: %v\n", err)
		//}

	case "bugzilla-download-comments":
		startTime := time.Now() // Record the start time

		// Call the download function
		operations.DownloadBugzillaComments(&bugzClient, db)

		// Calculate and print the time taken
		duration := time.Since(startTime)
		fmt.Printf("DownloadBugzillaComments completed in %s\n", duration)

	case "bugzilla-download-attachments":
		startTime := time.Now() // Record the start time

		// Call the download function
		operations.DownloadBugzillaAttachments(&bugzClient, db)

		// Calculate and print the time taken
		duration := time.Since(startTime)
		fmt.Printf("DownloadBugzillaAttachments completed in %s\n", duration)

	case "gitea-upload-bugs":

		var url string
		if len(os.Args) < 3 {
			url = forgejo.DEFAULT_GITEA_URL
			fmt.Println("using default gitea url", url)
		} else {
			url = os.Args[2]
		}

		// TODO fix this after gitea is decoupled from database
		//giteaClient := giteacustom.New(url)
		//giteaClient.UploadBugs(db.Conn)

	default:
		fmt.Println("invalid command")
		printHelp()
	}
}

func printHelp() { // not sure about functions descriptions
	fmt.Println("available commands:\n" +
		"bugzilla-download-bugs - Downloads bugs from bugzilla\n" +
		"bugzilla-show-bugs - Shows bugzilla bugs\n" +
		"bugzilla-download-users - Downloads users from bugzilla\n" +
		"bugzilla-download-comments - Downloads comments from bugs db. Can take long time ~16+ hours.\n" +
		"bugzilla-download-attachments - Downloads attachments from bugs db. Can take long time ~16+ hours.\n" +
		"gitea-upload-bugs [url] - Uploads bugs to gitea from local bugzilla db\n" +
		"help - shows available commands")

}
