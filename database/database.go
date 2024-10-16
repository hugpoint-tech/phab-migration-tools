package database

import (
	"context"
	"embed"
	"encoding/json"
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

func New(path string) *DB { // Return a pointer to DB
	var err error
	db := &DB{} // Create a pointer to the DB struct

	// Create a new connection pool
	db.pool, err = sqlitex.NewPool(path, sqlitex.PoolOptions{
		PoolSize: runtime.NumCPU() * 4,
	})
	util.CheckFatal("failed to create connection pool", err)

	// Get a connection from the pool to apply the schema
	conn := db.pool.Get(context.Background())
	if conn == nil {
		util.Fatal("failed to open database connection")
	}
	defer db.pool.Put(conn) // Put the connection back after applying the schema

	// Apply the schema
	err = sqlitex.ExecScript(conn, qSchema)
	util.CheckFatal("error applying schema", err)

	return db // Return a pointer to the DB struct
}

func (db *DB) GetDistinctUsers() ([]bugzilla.User, error) {
	var users []bugzilla.User

	conn := db.pool.Get(context.Background())
	if conn == nil {
		util.Fatal("failed to open database connection")
	}
	defer db.pool.Put(conn)

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

func (db *DB) InsertBug(bugs ...bugzilla.Bug) (err error) {
	conn := db.pool.Get(context.Background())
	if conn == nil {
		util.Fatal("failed to open database connection")
	}
	defer db.pool.Put(conn)
	var txCommit func(*error)

	txCommit, err = sqlitex.ImmediateTransaction(conn)
	if err != nil {
		return err
	}
	defer txCommit(&err)

	bugsJsons := make([]string, 0, len(bugs))
	for _, bug := range bugs {
		bugJson, err := json.Marshal(bug)
		if err != nil {
			return err
		}
		bugsJsons = append(bugsJsons, string(bugJson))
	}

	stmt := conn.Prep(qInsertBug)
	for i := range bugs {
		stmt.BindInt64(1, int64(bugs[i].ID))
		stmt.BindText(2, bugs[i].CreationTime)
		stmt.BindText(3, bugs[i].Creator)
		stmt.BindText(4, bugsJsons[i])

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

func (db *DB) InsertComment(comments ...bugzilla.Comment) (err error) {
	conn := db.pool.Get(context.Background())

	if conn == nil {
		util.Fatal("failed to open database connection")
	}
	defer db.pool.Put(conn)
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

func (db *DB) InsertAttachment(attachments ...bugzilla.Attachment) (err error) {
	conn := db.pool.Get(context.Background())
	if conn == nil {
		util.Fatal("failed to open database connection")
	}
	defer db.pool.Put(conn)

	var txCommit func(*error)

	txCommit, err = sqlitex.ImmediateTransaction(conn)
	if err != nil {
		return
	}
	defer txCommit(&err)

	stmt := conn.Prep(qInsertAttachments)

	for _, a := range attachments {
		stmt.BindInt64(1, int64(a.ID))
		stmt.BindInt64(2, int64(a.BugID))
		stmt.BindText(3, a.CreationTime)
		stmt.BindText(4, a.Creator)
		stmt.BindText(5, a.Summary)
		stmt.BindText(6, string(a.Data))

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

func (db *DB) ForEachBug(pred func(b bugzilla.Bug) error) error {
	conn := db.pool.Get(context.Background())
	if conn == nil {
		util.Fatal("failed to open database connection")
	}
	defer db.pool.Put(conn)

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
	defer db.pool.Put(conn)

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
