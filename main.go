package main

import (
	"fmt"
	"os"
	. "hugpoint.tech/freebsd/forge/bugz"
	. "hugpoint.tech/freebsd/forge/usersJson"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Not enough arguments. Use help")
		return
	}
	command := os.Args[1]

	switch command {
	case "bugzilla-download-bugs":
		DownloadBugs()
	case "help":
		printHelp()
	case "bugzilla-list-bugs":
		listBugs()
	case "bugzilla-download-users":
		GetUsersJson()
	case "bugzilla-show-bugs":
		showBugs()
	default:
		fmt.Println("Use help")
	}
}

func printHelp() { // not sure about functions descriptions
	fmt.Println("Available commands:\n" +
		"bugzilla-download-bugs - Downloads bugs from bugzilla\n" +
		"bugzilla-show-bugs - Shows bugzilla bugs\n" +
		"bugzilla-list-bugs - Displays downloaded bugs\n" +
		"bugzilla-download-users - Downloads users from bugzilla\n" +
		"help - Shows available commands")
}
func listBugs() {
	fmt.Println("Listing bugs")
}
func showBugs() {
	fmt.Println("Showing bugs")
}

