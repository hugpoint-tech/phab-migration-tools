package main

import (
	"fmt"
	"os"
	. "hugpoint.tech/freebsd/forge/bugz"
)

func main() {
	if len(os.Args) < 2 {
			fmt.Println("not enough arguments. use help")
			return
		}
		command := os.Args[1]

		switch command {
		case "bugzilla-download-bugs":
			DownloadBugzillaBugs()
		case "help":
			printHelp()
		case "bugzilla-list-bugs":
			listBugs()
		case "bugzilla-download-users":
			DownloadBugzillaUsers()
		case "bugzilla-show-bugs":
			showBugs()
		default:
			fmt.Println("use help")
		}
	}

	func printHelp() { // not sure about functions descriptions
		fmt.Println("available commands:\n" +
			"bugzilla-download-bugs - downloads bugs from bugzilla\n" +
			"bugzilla-show-bugs - shows bugzilla bugs\n" +
			"bugzilla-list-bugs - displays downloaded bugs\n" +
			"bugzilla-download-users - downloads users from bugzilla\n" +
			"help - shows available commands")
	}
	func listBugs() {
		fmt.Println("listing bugs")
	}

	func showBugs() {
		fmt.Println("showing bugs")
	}

	/*func getBugsFromBugzilla(pullConcurrencyLevel int, chanBug chan bugz.Bug) {
		tn := time.Now().UTC()
		fmt.Println("getting bugs from Bugzilla start", tn.String())
		defer func() { fmt.Println("getBugsFromBugzilla finished in", time.Since(tn)) }()

		chanDone := make(chan struct{})
		bugzilla := bugz.NewBugzClient()
		go bugzilla.Bugs().GetAll(chanDone, chanBug, pullConcurrencyLevel)

		go func() {
			<-chanDone
			fmt.Println("got bugs in", time.Since(tn))
			time.Sleep(1 * time.Minute) // wait a bit for all http clients read data and put it into chanBug
			close(chanBug)
		}()
	}

	func readBugsFromFiles(ctx context.Context, chanBug chan bugz.Bug) error {
		folder := "../bsdata/bugzilla-bugs/"
		files, err := os.ReadDir(folder)
		if err != nil {
			return err
		}
		for _, file := range files {
			dat, err := os.ReadFile(folder + file.Name())
			if err != nil {
				return err
			}
			var bug bugz.Bug
			if err := json.Unmarshal(dat, &bug); err != nil {
				return err
			}
			chanBug <- bug
		}

		return nil*/
	/* Create a new instance of BugzClient
	bc := new(BugzClient)

	// Call the DownloadAllBugs method
	//err := bc.DownloadAllBugs()
	token, _ := bc.GetToken()
	//if err != nil {
	// Handle error
	//	panic(err)
	//}
	println(token)
	// Download completed successfully
	//println("Download of bugs completed successfully.")*/
