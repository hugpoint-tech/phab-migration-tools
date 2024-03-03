package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("not enough arguments. use help")
		return
	}
	command := os.Args[1]

	switch command {
	case "bugzilla-download-bugs":
		downloadBugs()
	case "help":
		printHelp()
	case "bugzilla-list-bugs":
		listBugs()
	case "bugzilla-download-users":
		downloadUsers()
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
func downloadBugs() {
	fmt.Println("downloading bugs")
}
func listBugs() {
	fmt.Println("listing bugs")
}
func downloadUsers() {
	fmt.Println("downloading users")
}
func showBugs() {
	fmt.Println("showing bugs")
}

