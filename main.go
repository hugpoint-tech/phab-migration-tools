// vim: set expandtab ts=4 sw=4 sts=4 :
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"hugpoint.tech/freebsd/forge/bugz"
	"hugpoint.tech/freebsd/forge/phab"
	. "hugpoint.tech/freebsd/forge/util"
)

func main() {
	fmt.Println("Awesome FreeBSD Phabricator to Gitea/Forgejo Migrator 9000")

	db, err := sql.Open("sqlite3", "./forge.db")
	CheckFatal(err)
	defer db.Close()

	createTables(db)

	getDataFromBugzilla(db)
	// getPhabricatorData(db)
}

func getDataFromBugzilla(db *sql.DB) {
	fmt.Println("Getting data from Bugzilla")

	tn := time.Now().UTC()
	bugzilla := bugz.NewBugzClient()
	bugs := bugzilla.Bugs().GetAll()

	fmt.Println("got bugs in", len(bugs), time.Since(tn))

	tx, err := db.Begin()
	CheckFatal(err)

	for i, bug := range bugs {
		fmt.Printf("\rInserting bug into file: %d", i)
		writeFile(bug)
	}
	fmt.Println("finished saving into files", time.Since(tn))

	for i, bug := range bugs {
		fmt.Printf("\rInserting bug into sql: %d", i)

		// writeFile(bug)

		_, err = tx.Exec(`
	    	INSERT OR REPLACE INTO bugs (
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
	    	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	    `,
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
		)

		if err != nil {
			tx.Rollback()
			CheckFatal(err)
		}
	}
	fmt.Println()
	tx.Commit()

	fmt.Println("finished in", time.Since(tn))
}

func createTables(db *sql.DB) {
	_, err := db.Exec(`
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
        )
    `)
	CheckFatal(err)

	_, err = db.Exec(`
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
    		);
    	`)
	CheckFatal(err)

	_, err = db.Exec(`
    		CREATE TABLE IF NOT EXISTS bugz_users (
    			id INTEGER PRIMARY KEY,
    			email TEXT,
    			real_name TEXT,
    			name TEXT
    		);
    	`)
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

func writeFile(bugData bugz.Bug) {
	bytes, err := json.Marshal(bugData)
	CheckFatal(err)

	err = os.WriteFile(fmt.Sprintf("%d.json", bugData.ID), bytes, 0644)
	CheckFatal(err)
}
