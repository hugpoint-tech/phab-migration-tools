package main

import (
	"fmt"
	. "hugpoint.tech/freebsd/forge/bugzilla"
	"hugpoint.tech/freebsd/forge/database"
	"hugpoint.tech/freebsd/forge/forgejo"
	"hugpoint.tech/freebsd/forge/operations"
	"os"
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
		operations.DownloadBugzillaBugs(&bugzClient, &db)
	case "help":
		printHelp()

	case "bugzilla-download-users":
		//err := bugzClient.DownloadBugzillaUsers()
		//if err != nil {
		//	log.Fatalf("Error downloading users: %v\n", err)
		//}

	case "bugzilla-download-comments":
		operations.DownloadBugzillaComments(&bugzClient, &db)

	case "bugzilla-download-attachments":
		operations.DownloadBugzillaAttachments(&bugzClient, &db)

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
		"bugzilla-download-bugs - downloads bugs from bugzilla\n" +
		"bugzilla-show-bugs - shows bugzilla bugs\n" +
		"bugzilla-download-users - downloads users from bugzilla\n" +
		"bugzilla-download-comments - downloads comments from bugs db\n" +
		"bugzilla-download-attachments - downloads attachments from bugs db\n" +
		"gitea-upload-bugs [url] - upload bugs to gitea from local bugzilla db\n" +
		"help - shows available commands")

}
