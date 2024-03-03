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

func printHelp() {
	fmt.Println("display available commands")
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

