package main

import (
	"fmt"
	. "hugpoint.tech/freebsd/forge/bugzilla"
	"hugpoint.tech/freebsd/forge/common/bugzilla"
	"hugpoint.tech/freebsd/forge/database"
	"hugpoint.tech/freebsd/forge/forgejo"
	"hugpoint.tech/freebsd/forge/util"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("not enough arguments. use help")
		return
	}
	command := os.Args[1]
	db := database.New("migrator.db")
	bc := NewBugzClient()

	switch command {
	case "bugzilla-download-bugs":

		//err := bc.DownloadBugzillaBugs()
		//if err != nil {
		//	log.Fatalf("Error downloading bugs: %v\n", err)
		//}
	case "help":
		printHelp()
	case "bugzilla-list-bugs":
		err := bc.ListBugs()
		if err != nil {
			log.Fatalf("Error listing bugs: %v\n", err)
		}
	case "bugzilla-download-users":
		//err := bc.DownloadBugzillaUsers()
		//if err != nil {
		//	log.Fatalf("Error downloading users: %v\n", err)
		//}
	case "bugzilla-show-bugs":
		if err := bc.ShowBugs(); err != nil {
			log.Fatalf("error showing bugs: %v\n", err)
		}
	case "bugzilla-download-comments":
		// Fetch bugs from the SQLite database
		err := db.ForEachBug(func(bug bugzilla.Bug) error {
			fmt.Println("Downloading comments for", bug.ID)
			// err := bc.DownloadAllComments()
			return nil
		})
		util.CheckFatal("failed to download comments", err)

	case "bugzilla-download-attachments":
		// Fetch bugs from the SQLite database
		//bugs, err := ListBugs(db.Conn)
		//if err != nil {
		//	log.Fatalf("Error fetching bugs: %v", err)
		//}
		//// Download attachments for each bug
		//for _, bug := range bugs {
		//	err := bc.DownloadBugAttachments(int64(bug.ID))
		//	if err != nil {
		//		fmt.Printf("Error downloading attachments for bug %d: %v", bug.ID, err)
		//	}
		//}
		//fmt.Println("Downloaded attachments for all bugs successfully.")
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
		"bugzilla-list-bugs - displays downloaded bugs\n" +
		"bugzilla-download-users - downloads users from bugzilla\n" +
		"bugzilla-download-comments - downloads comments from bugs db\n" +
		"bugzilla-download-attachments - downloads attachments from bugs db\n" +
		"gitea-upload-bugs [url] - upload bugs to gitea from local bugzilla db\n" +
		"help - shows available commands")

}
