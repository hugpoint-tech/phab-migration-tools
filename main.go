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

	return nil
}

func getBugsFromBugzilla(pullConcurrencyLevel int, chanBug chan bugz.Bug) {
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
