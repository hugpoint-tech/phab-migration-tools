package main

import (
	"context"
	"encoding/json"
	"fmt"
	"hugpoint.tech/freebsd/forge/bugz"
	"hugpoint.tech/freebsd/forge/database"
	"os"
	"os/signal"
	"time"

	. "hugpoint.tech/freebsd/forge/util"
)

func oldmain() {
	fmt.Println("Awesome FreeBSD Phabricator to Gitea/Forgejo Migrator 9000")
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	tn := time.Now().UTC()
	defer func() { fmt.Println("main finished in", time.Since(tn)) }()

	db, err := database.New(ctx, "./zombiezen.db")
	CheckFatal("database.New", err)
	totalBugsInDB, err := db.GetBugsTotal(ctx)
	fmt.Println("totalBugsInDB", totalBugsInDB)

	CheckFatal("db.GetTotalBugs", err)
	totalUsersInDB, err := db.GetUsersTotal(ctx)
	CheckFatal("db.GetUsersTotal", err)
	fmt.Println("totalUsersInDB", totalUsersInDB)

	maxUserID, err := db.GetUsersMaxID(ctx)
	CheckFatal("db.GetUsersMaxID", err)
	fmt.Println("maxUserID", maxUserID)

	// if totalBugsInDB == 0 {
	// 	pullConcurrencyLevel := 5
	// 	chanBug := make(chan bugz.Bug, 300*1000)
	// 	getBugsFromBugzilla(pullConcurrencyLevel, chanBug)
	// 	err = db.CreateBugs(ctx, chanBug)
	// 	CheckFatal("db.CreateBugs", err)
	// }

	// {
	// 	chanBug := make(chan bugz.Bug, 300*1000)
	// 	go func() {
	// 		err = db.CreateBugs(ctx, chanBug)
	// 		CheckFatal("db.CreateBugs", err)
	// 	}()

	// 	err = readBugsFromFiles(ctx, chanBug)
	// 	CheckFatal("readBugsFromFiles", err)
	// 	time.Sleep(11 * time.Second)
	// 	close(chanBug)
	// }

	//gc, err := git.NewClient("http://localhost:3000/")
	//CheckFatal("git.NewClient", err)

	{ // users block
		chanUsrs := make(chan bugz.User, 100*1000)

		go func() {
			err = db.GetUsersFromBugs(ctx, chanUsrs)
			CheckFatal("db.GetUsersFromBugs", err)
			time.Sleep(11 * time.Second)
			close(chanUsrs)
		}()

		err = db.CreateUsers(ctx, chanUsrs)
		CheckFatal("db.CreateUsers", err)

		users, err := db.GetUsers(ctx) // https://bugzilla.readthedocs.io/en/latest/api/core/v1/user.html
		CheckFatal("db.GetUsers", err)
		fmt.Println("got users from db from bugs", len(users))

		{ // gitea block - create users
			//	err = gc.DeleteAll(users)
			CheckFatal("DeleteAll", err)

			//	err = gc.CreateUsers(users)
			//CheckFatal("gc.CreateUsers", err)
		}
	}

	{ // bugs block
		bugsChan := make(chan bugz.Bug, 300*1000)
		go func() {
			err = db.GetBugs(ctx, bugsChan)
			CheckFatal("db.GetBugs", err)
			time.Sleep(11 * time.Second)
			close(bugsChan)
		}()
		{ // gitea block
			//	err = gc.CreateBugz(bugsChan)
			//	CheckFatal("gc.CreateBugz", err)
		}
	}
	//err = db.GetPhabricatorData(ctx)
	//CheckFatal("getPhabricatorData", err)
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
