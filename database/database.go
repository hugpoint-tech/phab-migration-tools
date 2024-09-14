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

func New(path string) DB { // Return a pointer to database
	var err error
	var db DB

	db.pool, err = sqlitex.NewPool(path, sqlitex.PoolOptions{
		PoolSize: runtime.NumCPU() * 4,
	})
	util.CheckFatal("failed to create connection pool", err)

	conn := db.pool.Get(context.Background())
	if conn == nil {
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
	if conn == nil {
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

func (db *DB) InsertBug(bug bugzilla.Bug) error {
	bugJson, err := json.Marshal(bug)
	util.CheckFatal("error marshalling bug JSON", err)

	conn := db.pool.Get(context.Background())
	if conn == nil {
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

	return sqlitex.ExecuteTransient(conn, qInsertBug, &execOptions)
}

func (db *DB) InsertComment(comments ...bugzilla.Comment) (err error) {
	conn := db.pool.Get(context.Background())
	if conn == nil {
		util.Fatal("failed to open database connection")
	}
	defer conn.Close()
	var txCommit func(*error)

	txCommit, err = sqlitex.ImmediateTransaction(conn)
	if err != nil {
		return
	}
	defer txCommit(&err)

	stmt := conn.Prep(qInsertComments)
	for _, c := range comments {
		stmt.BindInt64(1, int64(c.ID))
		stmt.BindInt64(2, int64(c.BugID))
		stmt.BindInt64(3, int64(c.AttachmentID))
		stmt.BindText(4, c.CreationTime)
		stmt.BindText(5, c.Creator)
		stmt.BindText(6, c.Text)

		_, err = stmt.Step()
		if err != nil {
			return
		}
		err = stmt.Reset()
		if err != nil {
			return
		}
	}

	return
}

func (db *DB) InsertAttachment(attachment bugzilla.Attachment) error {
	conn := db.pool.Get(context.Background())
	if conn == nil {
		util.Fatal("failed to open database connection")
	}
	defer conn.Close()

	execOptions := sqlitex.ExecOptions{
		Args: []interface{}{
			attachment.ID,
			attachment.BugID,
			attachment.CreationTime,
			attachment.Creator,
			attachment.Summary,
			attachment.Data,
		},
	}

	// Try inserting the comment into the database
	err := sqlitex.Execute(conn, qInsertAttachments, &execOptions)
	if err != nil {
		return fmt.Errorf("error inserting attachments: %v", err)
	}

	return nil
}

func (db *DB) ForEachBug(pred func(b bugzilla.Bug) error) error {
	conn := db.pool.Get(context.Background())
	if conn == nil {
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
	if conn == nil {
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
