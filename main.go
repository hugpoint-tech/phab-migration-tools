// vim: set expandtab ts=4 sw=4 sts=4 :
package main

import (
	"database/sql"
	"fmt"
	"hugpoint.tech/freebsd/forge/bugz"
	"hugpoint.tech/freebsd/forge/phab"
	. "hugpoint.tech/freebsd/forge/util"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func main() {

	fmt.Println("Awesome FreeBSD Phabricator to Gitea/Forgejo Migrator 9000")

	db, err := sql.Open("sqlite3", "./forge.db")
	CheckFatal(err)
	defer db.Close()

	_, err = db.Exec(`
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

	_, err = db.Exec(`
    		CREATE TABLE IF NOT EXISTS bugz_users (
    			id INTEGER PRIMARY KEY,
    			email TEXT,
    			real_name TEXT,
    			name TEXT
    		);
    	`)

	CheckFatal(err)

	fmt.Println("Getting data from Bugzilla")
	bugzilla := bugz.NewBugzClient()
	bugs := bugzilla.Bugs().GetAll()

	tx, err := db.Begin()
	CheckFatal(err)

	for i, bug := range bugs {

		fmt.Printf("\rInserting bug into sql: %d", i)
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

	fmt.Println("Getting data from Phabricator")
	phab := phab.NewPhabClient()
	users := phab.UserSearch().Get()

	// Begin a transaction
	tx, err = db.Begin()
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
	}
	fmt.Println()
	tx.Commit()
}
