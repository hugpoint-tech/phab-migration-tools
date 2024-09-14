package database

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"hugpoint.tech/freebsd/forge/common/bugzilla"
	"hugpoint.tech/freebsd/forge/util"
	"log"
	"runtime"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

//go:embed *.sql
var schemaFS embed.FS

type DB struct {
	pool *sqlitex.Pool

	QInsertBug         string
	QSelectBugs        string
	QInsertComments    string
	QInsertUsers       string
	QInsertAttachments string
	QDistinctUsers     string
}

func New(path string) DB { // Return a pointer to DB
	var err error
	var sql []byte
	var result DB

	result.pool, err = sqlitex.NewPool(path, sqlitex.PoolOptions{
		PoolSize: runtime.NumCPU() * 4,
	})

	conn := result.pool.Get(context.Background())
	if conn != nil {
		util.Fatal("failed to open database connection")
	}
	defer conn.Close()

	util.CheckFatal("error opening database", err)
	sql, err = schemaFS.ReadFile("schema.sql")
	util.CheckFatal("failed to read schema", err)

	err = sqlitex.ExecScript(conn, string(sql))
	util.CheckFatal("error applying schema", err)

	sql, err = schemaFS.ReadFile("insert.sql")
	util.CheckFatal("failed to read embedded file", err)
	result.QInsertBug = string(sql)

	sql, err = schemaFS.ReadFile("select_all_bugs.sql")
	util.CheckFatal("failed to read embedded file", err)
	result.QSelectBugs = string(sql)

	sql, err = schemaFS.ReadFile("distinct.sql")
	util.CheckFatal("failed to read embedded file", err)
	result.QDistinctUsers = string(sql)

	sql, err = schemaFS.ReadFile("insert_comments.sql")
	util.CheckFatal("failed to read embedded file", err)
	result.QInsertComments = string(sql)

	sql, err = schemaFS.ReadFile("insert_user.sql")
	util.CheckFatal("failed to read embedded file", err)
	result.QInsertUsers = string(sql)

	sql, err = schemaFS.ReadFile("insert_attachments.sql")
	util.CheckFatal("failed to read embedded file", err)
	result.QInsertAttachments = string(sql)

	return result // Return the pointer to DB
}

func (db *DB) GetDistinctUsers() ([]bugzilla.User, error) {
	var users []bugzilla.User

	conn := db.pool.Get(context.Background())
	if conn != nil {
		util.Fatal("failed to open database connection")
	}
	defer conn.Close()

	execOptions := &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			user := bugzilla.User{
				Email:    stmt.ColumnText(0),
				Name:     stmt.ColumnText(1),
				RealName: stmt.ColumnText(2),
			}
			if user.Email == "" && user.Name == "" && user.RealName == "" {
				log.Printf("Empty user detected: %+v", user)
			} else {
				users = append(users, user)
			}
			return nil
		},
	}

	// Execute distinct query on the bugs table to retrieve unique user data
	err := sqlitex.ExecuteTransient(conn, db.QDistinctUsers, execOptions)
	if err != nil {
		return nil, err
	}

	return users, nil

}

func (db *DB) InsertBug(bug bugzilla.Bug) {
	bugJson, err := json.Marshal(bug)
	util.CheckFatal("error marshalling bug JSON", err)

	conn := db.pool.Get(context.Background())
	if conn != nil {
		util.Fatal("failed to open database connection")
	}
	defer conn.Close()

	execOptions := sqlitex.ExecOptions{
		Args: []interface{}{
			bug.ID,
			bug.CreationTime,
			bug.Creator,
			string(bugJson),
		},
	}

	err = sqlitex.ExecuteTransient(conn, db.QInsertBug, &execOptions)
	util.CheckFatal(fmt.Sprintf("error inserting bug %d", bug.ID), err)
}

func (db *DB) InsertComment(comment bugzilla.Comment) error {
	conn := db.pool.Get(context.Background())
	if conn != nil {
		util.Fatal("failed to open database connection")
	}
	defer conn.Close()

	execOptions := sqlitex.ExecOptions{
		Args: []interface{}{
			comment.ID,
			comment.BugID,
			comment.AttachmentID,
			comment.CreationTime,
			comment.Creator,
			comment.Text,
		},
	}

	// Try inserting the comment into the database
	err := sqlitex.Execute(conn, db.QInsertComments, &execOptions)
	if err != nil {
		return fmt.Errorf("error inserting comment: %v", err)
	}

	return nil
}

func (db *DB) ForEachBug(pred func(b bugzilla.Bug) error) error {
	conn := db.pool.Get(context.Background())
	if conn != nil {
		util.Fatal("failed to open database connection")
	}
	defer conn.Close()

	opts := sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			txt := stmt.ColumnText(0)
			var bug bugzilla.Bug
			err := json.Unmarshal([]byte(txt), &bug)
			if err != nil {
				return err
			}
			return pred(bug)
		},
		Args: make([]any, 0),
	}

	return sqlitex.Execute(conn, db.QSelectBugs, &opts)
}
