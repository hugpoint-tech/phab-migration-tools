// vim: set expandtab ts=4 sw=4 sts=4 :
package main

import (
	"database/sql"
	"fmt"
	"runtime"
	"strings"
	"time"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"

	"hugpoint.tech/freebsd/forge/bugz"
	"hugpoint.tech/freebsd/forge/phab"
	. "hugpoint.tech/freebsd/forge/util"
)

func main() {
	fmt.Println("Awesome FreeBSD Phabricator to Gitea/Forgejo Migrator 9000")
	tn := time.Now().UTC()
	defer func() { fmt.Println("main finished in", time.Since(tn)) }()

	conn, err := sqlite.OpenConn("./zombiezen.db", sqlite.OpenReadWrite, sqlite.OpenCreate)
	CheckFatal(err)
	defer conn.Close()

	err = sqlitex.ExecuteTransient(conn, "SELECT count(id) from bugs;", &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			fmt.Println("before search totalBugs", stmt.ColumnText(0))
			return nil
		},
	})
	CheckFatal(err)

	createTables(conn)
	pullConcurrencyLevel := 5
	getDataFromBugzilla(conn, pullConcurrencyLevel)

	fmt.Println("getDataFromBugzilla finished")
	printMemUsage()

	err = sqlitex.ExecuteTransient(conn, "SELECT count(id) from bugs;", &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			fmt.Println("totalBugs", stmt.ColumnText(0))
			return nil
		},
	})
	CheckFatal(err)

	// getPhabricatorData(db)
}

func getDataFromBugzilla(conn *sqlite.Conn, pullConcurrencyLevel int) {
	printMemUsage()

	tn := time.Now().UTC()
	fmt.Println("Getting data from Bugzilla start", tn.String())
	defer func() { fmt.Println("finished in", time.Since(tn)) }()

	chanDone := make(chan struct{})
	chanBug := make(chan bugz.Bug, 300*1000)
	bugzilla := bugz.NewBugzClient()
	go bugzilla.Bugs().GetAll(chanDone, chanBug, pullConcurrencyLevel)

	go func() {
		<-chanDone
		fmt.Println("got bugs in", time.Since(tn))
		time.Sleep(1 * time.Minute) // wait a bit for all http clients read data and put it into chanBug
		close(chanBug)
		printMemUsage()
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

func getPhabricatorData(db *sql.DB) {
	fmt.Println("Getting data from Phabricator")
	phab := phab.NewPhabClient()
	users := phab.UserSearch().Get()

	tx, err := db.Begin() // Begin a transaction
	CheckFatal(err)
	stmt, err := tx.Prepare(`
    INSERT INTO phab_users (
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
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`)
	CheckFatal(err)
	defer stmt.Close()

	for i, user := range users {
		fmt.Printf("\rInserting phab user into sql: %d", i)
		_, err = stmt.Exec(
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
		)
		if err != nil {
			fmt.Printf("Inserting phab user into sql %v, err: %s", i, err)
		}
	}
	fmt.Println()
	tx.Commit()
}

func printMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("\nAlloc = %v MiB", m.Alloc/1024/1024)
	fmt.Printf("\tTotalAlloc = %v MiB", m.TotalAlloc/1024/1024)
	fmt.Printf("\tSys = %v MiB", m.Sys/1024/1024)
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}
