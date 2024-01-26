// vim: set expandtab ts=4 sw=4 sts=4 :
package main

import (
	"context"
	"fmt"
	"time"

	"hugpoint.tech/freebsd/forge/bugz"
	"hugpoint.tech/freebsd/forge/database"
	"hugpoint.tech/freebsd/forge/git"
	. "hugpoint.tech/freebsd/forge/util"
)

func main() {
	fmt.Println("Awesome FreeBSD Phabricator to Gitea/Forgejo Migrator 9000")
	ctx := context.Background()

	tn := time.Now().UTC()
	defer func() { fmt.Println("main finished in", time.Since(tn)) }()

	db, err := database.New(ctx, "./zombiezen.db")
	CheckFatal("database.New", err)
	totalBugsInDB, err := db.GetBugsTotal(ctx)
	CheckFatal("db.GetTotalBugs", err)
	totalUsersInDB, err := db.GetUsersTotal(ctx)
	CheckFatal("db.GetUsersTotal", err)
	fmt.Println("totalUsersInDB", totalUsersInDB)

	if totalBugsInDB == 0 {
		pullConcurrencyLevel := 5
		chanBug := make(chan bugz.Bug, 300*1000)
		getBugsFromBugzilla(pullConcurrencyLevel, chanBug)
		err = db.CreateBugs(ctx, chanBug)
		CheckFatal("db.CreateBugs", err)
	}

	gc, err := git.NewClient("http://localhost:3000/")
	CheckFatal("git.NewClient", err)

	{ // users block
		chanUsrs := make(chan bugz.User, 100*1000)
		getUsersFromBugzilla(chanUsrs)
		err = db.CreateUsers(ctx, chanUsrs)
		CheckFatal("db.CreateUsers", err)

		users, err := db.GetUsers(ctx) // https://bugzilla.readthedocs.io/en/latest/api/core/v1/user.html
		CheckFatal("db.GetUsers", err)
		fmt.Println("db.GetUsers users", len(users))

		// { // gitea block - create users
		// 	err = gc.CreateUsers(users)
		// 	CheckFatal("gc.CreateUsers", err)
		// }
	}

	if false { // bugs block
		bugsChan := make(chan bugz.Bug, 300*1000)
		go func() {
			err = db.GetBugs(ctx, bugsChan)
			CheckFatal("db.GetBugs", err)
			time.Sleep(1 * time.Minute)
			close(bugsChan)
		}()
		{ // gitea block
			err = gc.CreateBugz(bugsChan)
			CheckFatal("gc.CreateBugz", err)
		}
	}
	// err = db.GetPhabricatorData(ctx)
	// CheckFatal("getPhabricatorData", err)
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

func getUsersFromBugzilla(chanUsrs chan bugz.User) {
	tn := time.Now().UTC()
	fmt.Println("getting users from Bugzilla start", tn.String())
	defer func() { fmt.Println("getUsersFromBugzilla finished in", time.Since(tn)) }()

	chanDone := make(chan struct{})
	bugzilla := bugz.NewBugzClient()
	go func() {
		err := bugzilla.UserAPI().Get(chanDone, chanUsrs)
		CheckFatal("bugzilla.UserAPI().Get()", err)
	}()

	<-chanDone
	fmt.Println("got bugs in", time.Since(tn))
	time.Sleep(1 * time.Second) // wait a bit
	close(chanUsrs)
}
