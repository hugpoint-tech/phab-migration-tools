// vim: set expandtab ts=4 sw=4 sts=4 :
package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"

	"hugpoint.tech/freebsd/forge/bugz"
	"hugpoint.tech/freebsd/forge/database"
	"hugpoint.tech/freebsd/forge/git"
	"hugpoint.tech/freebsd/forge/phab"
	. "hugpoint.tech/freebsd/forge/util"
)

func main() {
	fmt.Println("Awesome FreeBSD Phabricator to Gitea/Forgejo Migrator 9000")
	ctx := context.Background()

	tn := time.Now().UTC()
	defer func() { fmt.Println("main finished in", time.Since(tn)) }()

	db, err := database.New(ctx, "./zombiezen.db")
	CheckFatal("database.New", err)
	totalBugsInDB, err := db.GetTotalBugs(ctx)
	CheckFatal("db.GetTotalBugs", err)

	if totalBugsInDB == 0 {
		pullConcurrencyLevel := 5
		chanBug := make(chan bugz.Bug, 300*1000)
		getDataFromBugzilla(pullConcurrencyLevel, chanBug)
		err = db.CreateBugs(ctx, chanBug)
		CheckFatal("db.CreateBugs", err)
	}

	gc, err := git.NewClient("http://localhost:3000/")
	CheckFatal("git.NewClient", err)

	{ // users block
		bugzilla := bugz.NewBugzClient()
		chanUsrs := make(chan bugz.User, 100*1000)

		err = bugzilla.UserAPI().Get(chanUsrs)
		CheckFatal("bugzilla.UserAPI().Get()", err)

		go func() {
			err = db.CreateUsers(ctx, chanUsrs)
			CheckFatal("db.CreateUsers", err)
			time.Sleep(1 * time.Minute)
			close(chanUsrs)
		}()

		users, err := db.GetUsers(ctx) // https://bugzilla.readthedocs.io/en/latest/api/core/v1/user.html
		CheckFatal("db.GetUsers", err)
		fmt.Println("db.GetUsers users", len(users))

		{ // gitea block - create users
			err = gc.CreateUsers(users)
			CheckFatal("gc.CreateUsers", err)
		}
	}

	{ // bugs block
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
	// getPhabricatorData(db)
}

func getDataFromBugzilla(pullConcurrencyLevel int, chanBug chan bugz.Bug) {
	tn := time.Now().UTC()
	fmt.Println("Getting data from Bugzilla start", tn.String())
	defer func() { fmt.Println("getDataFromBugzilla finished in", time.Since(tn)) }()

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

func getPhabricatorData(conn *sqlite.Conn) {
	fmt.Println("Getting data from Phabricator")
	phab := phab.NewPhabClient()
	users := phab.UserSearch().Get()

	for i, user := range users {
		fmt.Printf("\rInserting phab user into sql: %d", i)

		err := sqlitex.ExecuteTransient(
			conn,
			`  INSERT INTO phab_users (
				id,
				type,
				phid,
				username,
				real_name,
				date_created,
				date_modified,
				roles,
				policy_edit,
				policy_view
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`,
			&sqlitex.ExecOptions{
				Args: []any{
					user.ID,
					user.Type,
					user.Phid,
					user.Fields.Username,
					user.Fields.RealName,
					user.Fields.DateCreated,
					user.Fields.DateModified,
					strings.Join(user.Fields.Roles, ","),
					user.Fields.Policy.Edit,
					user.Fields.Policy.View,
				},
			},
		)
		CheckFatal("getPhabricatorData", err)
	}
}
