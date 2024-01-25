// vim: set expandtab ts=4 sw=4 sts=4 :
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"code.gitea.io/sdk/gitea"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"

	"hugpoint.tech/freebsd/forge/bugz"
	"hugpoint.tech/freebsd/forge/phab"
	. "hugpoint.tech/freebsd/forge/util"
)

var dbpool *sqlitex.Pool

func main() {
	fmt.Println("Awesome FreeBSD Phabricator to Gitea/Forgejo Migrator 9000")
	ctx := context.Background()

	tn := time.Now().UTC()
	defer func() { fmt.Println("main finished in", time.Since(tn)) }()

	var err error
	dbpool, err = sqlitex.NewPool("./zombiezen.db", sqlitex.PoolOptions{
		PoolSize: 10,
	})
	CheckFatal(err)

	conn := dbpool.Get(ctx)
	if conn == nil {
		fmt.Println("got nil conn")
		return
	}
	defer dbpool.Put(conn)

	var totalBugsInDB int
	err = sqlitex.ExecuteTransient(conn, "SELECT count(id) from bugs;", &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			totalBugsInDB = stmt.ColumnInt(0)
			fmt.Println("before search totalBugsInDB", totalBugsInDB)
			return nil
		},
	})
	CheckFatal(err)

	if totalBugsInDB == 0 {
		createTables(conn)
		pullConcurrencyLevel := 5
		getDataFromBugzilla(conn, pullConcurrencyLevel)
	}

	err = sqlitex.ExecuteTransient(conn, "SELECT count(id) from bugs;", &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			fmt.Println("totalBugs", stmt.ColumnText(0))
			return nil
		},
	})
	CheckFatal(err)

	stmt := conn.Prep("SELECT id FROM bugs WHERE id = $id;")
	stmt.SetText("$id", "1")
	for {
		if hasRow, err := stmt.Step(); err != nil {
			CheckFatal(err)
		} else if !hasRow {
			break
		}
		foo := stmt.GetText("id")
		fmt.Println("id", foo)
	}

	giteaLogin, ok := os.LookupEnv("GITEA_LOGIN")
	if !ok {
		panic("GITEA_LOGIN is not set")
	}
	giteaPassword, ok := os.LookupEnv("GITEA_PASSWORD")
	if !ok {
		panic("GITEA_PASSWORD is not set")
	}

	gc, err := gitea.NewClient("http://localhost:3000/", gitea.SetBasicAuth(giteaLogin, giteaPassword))
	CheckFatal(err)

	existingTokens, resp, err := gc.ListAccessTokens(gitea.ListAccessTokensOptions{})
	CheckFatal(err)

	fmt.Println("existingTokens", existingTokens)
	fmt.Println("existingTokens resp", resp)
	if len(existingTokens) > 0 {
		fmt.Println("existingTokens 0 - ", existingTokens[0])
	} else {
		tkn, resp, err := gc.CreateAccessToken(gitea.CreateAccessTokenOption{
			Name: "my_tkn",
			Scopes: []gitea.AccessTokenScope{
				gitea.AccessTokenScopeAll,
			},
		})
		CheckFatal(err)

		// if err != nil {
		// 	fmt.Println("tkn err", err) // on repetitive calls don't fail
		// }
		fmt.Println("tkn", tkn)
		fmt.Println("tkn resp", resp)
	}

	repo, resp, err := gc.CreateRepo(gitea.CreateRepoOption{
		Name:          "testrepo",
		Description:   "",
		Private:       false,
		IssueLabels:   "",
		AutoInit:      false,
		Template:      false,
		Gitignores:    "",
		License:       "",
		Readme:        "",
		DefaultBranch: "",
		TrustModel:    "",
	})
	CheckFatal(err)
	fmt.Println("repo", repo)
	fmt.Println("repo resp", resp)

	iss, resp, err := gc.CreateIssue(repo.Owner.UserName, repo.Name, gitea.CreateIssueOption{
		Title:     "test_title",
		Body:      "",
		Ref:       "",
		Assignees: []string{},
		Deadline:  &time.Time{},
		Milestone: 0,
		Labels:    []int64{},
		Closed:    false,
	})
	CheckFatal(err)
	fmt.Println("iss", iss)
	fmt.Println("iss resp", resp)

	// getPhabricatorData(db)
}

func getDataFromBugzilla(conn *sqlite.Conn, pullConcurrencyLevel int) {
	tn := time.Now().UTC()
	fmt.Println("Getting data from Bugzilla start", tn.String())
	defer func() { fmt.Println("getDataFromBugzilla finished in", time.Since(tn)) }()

	chanDone := make(chan struct{})
	chanBug := make(chan bugz.Bug, 300*1000)
	bugzilla := bugz.NewBugzClient()
	go bugzilla.Bugs().GetAll(chanDone, chanBug, pullConcurrencyLevel)

	go func() {
		<-chanDone
		fmt.Println("got bugs in", time.Since(tn))
		time.Sleep(1 * time.Minute) // wait a bit for all http clients read data and put it into chanBug
		close(chanBug)
	}()

	var i int
	for bug := range chanBug {
		i++
		fmt.Printf("\rInserting bug into sql: %d", i)

		err := sqlitex.ExecuteTransient(
			conn,
			`INSERT OR REPLACE INTO bugs (
				id,
				product,
				status,
				priority,
				severity,
				component,
				platform,
				is_cc_accessible,
				creation_time,
				resolution,
				is_creator_accessible,
				version,
				summary,
				url,
				is_confirmed,
				last_change_time,
				dupe_of,
				creator,
				assigned_to
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`,
			&sqlitex.ExecOptions{
				Args: []any{
					bug.ID,
					bug.Product,
					bug.Status,
					bug.Priority,
					bug.Severity,
					bug.Component,
					bug.Platform,
					bug.IsCCAccessible,
					bug.CreationTime,
					bug.Resolution,
					bug.IsCreatorAccessible,
					bug.Version,
					bug.Summary,
					bug.URL,
					bug.IsConfirmed,
					bug.LastChangeTime,
					bug.DupeOf,
					bug.Creator,
					bug.AssignedTo,
				},
			},
		)
		CheckFatal(err)
	}
}

func createTables(conn *sqlite.Conn) {
	err := sqlitex.ExecuteTransient(conn, `
	CREATE TABLE IF NOT EXISTS phab_users (
		id              INTEGER PRIMARY KEY,
		type            TEXT,
		phid            TEXT,
		username        TEXT,
		real_name       TEXT,
		date_created    INTEGER,
		date_modified   INTEGER,
		roles           TEXT,
		policy_edit     TEXT,
		policy_view     TEXT
	);`, nil)
	CheckFatal(err)

	err = sqlitex.ExecuteTransient(conn, `
	CREATE TABLE IF NOT EXISTS bugs (
		id INTEGER PRIMARY KEY,
		product TEXT,
		status TEXT,
		priority TEXT,
		severity TEXT,
		component TEXT,
		platform TEXT,
		is_cc_accessible INTEGER,
		creation_time TEXT,
		resolution TEXT,
		is_open INTEGER,
		is_creator_accessible INTEGER,
		version TEXT,
		summary TEXT,
		url TEXT,
		is_confirmed INTEGER,
		last_change_time TEXT,
		dupe_of INTEGER,
		creator TEXT,
		assigned_to TEXT
	);`, nil)
	CheckFatal(err)

	err = sqlitex.ExecuteTransient(conn, `
	CREATE TABLE IF NOT EXISTS bugz_users (
		id INTEGER PRIMARY KEY,
		email TEXT,
		real_name TEXT,
		name TEXT
	);`, nil)
	CheckFatal(err)

	err = sqlitex.ExecuteTransient(conn, `
    		CREATE TABLE IF NOT EXISTS bugz_users (
    			id INTEGER PRIMARY KEY,
    			email TEXT,
    			real_name TEXT,
    			name TEXT
    		);`, nil)
	CheckFatal(err)
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
		CheckFatal(err)
	}
}
