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

var (
	//go:embed *.sql
	schemaFS embed.FS

	qSchema            = embeddedSQL("schema.sql")
	qInsertBug         = embeddedSQL("insert.sql")
	qSelectBugs        = embeddedSQL("select_all_bugs.sql")
	qSelectUsers       = embeddedSQL("select_all_users.sql")
	qDistinctUsers     = embeddedSQL("select_distinct_users.sql")
	qInsertComments    = embeddedSQL("insert_comments.sql")
	qInsertUsers       = embeddedSQL("insert_user.sql")
	qInsertAttachments = embeddedSQL("insert_attachments.sql")
)

type DB struct {
	pool *sqlitex.Pool
}

func embeddedSQL(filename string) string {
	sql, err := schemaFS.ReadFile(filename)
	util.CheckFatal("failed to read embedded SQL file", err)
	return string(sql)
}

func New(path string) DB { // Return a pointer to DB
	var err error
	var db DB

	db.pool, err = sqlitex.NewPool(path, sqlitex.PoolOptions{
		PoolSize: runtime.NumCPU() * 4,
	})

	conn := db.pool.Get(context.Background())
	if conn != nil {
		util.Fatal("failed to open database connection")
	}
	defer conn.Close()
	util.CheckFatal("error opening database", err)

	err = sqlitex.ExecScript(conn, qSchema)
	util.CheckFatal("error applying schema", err)

	return db
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
	err := sqlitex.ExecuteTransient(conn, qDistinctUsers, execOptions)
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

	err = sqlitex.ExecuteTransient(conn, qInsertBug, &execOptions)
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
	err := sqlitex.Execute(conn, qInsertComments, &execOptions)
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

	return sqlitex.Execute(conn, qSelectBugs, &opts)
}

func (db *DB) ForEachUser(pred func(b bugzilla.User) error) error {
	conn := db.pool.Get(context.Background())
	if conn != nil {
		util.Fatal("failed to open database connection")
	}
	defer conn.Close()

	opts := sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			var user bugzilla.User
			user.Email = stmt.ColumnText(0)
			user.Name = stmt.ColumnText(1)

			return pred(user)
		},
		Args: make([]any, 0),
	}

	return sqlitex.Execute(conn, qSelectUsers, &opts)
}
