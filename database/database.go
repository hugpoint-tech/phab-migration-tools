package database

import (
	"embed"
	"encoding/json"
	"hugpoint.tech/freebsd/forge/common/bugzilla"
	"hugpoint.tech/freebsd/forge/util"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

//go:embed *.sql
var schemaFS embed.FS

type DB struct {
	Conn *sqlite.Conn

	QInsertBug         string
	QInsertComments    string
	QInsertUsers       string
	QInsertAttachments string
	QDistinctUsers     string
}

func New(path string) DB {
	var err error
	var sql []byte
	var result DB

	result.Conn, err = sqlite.OpenConn(path, 0)
	util.CheckFatal("error opening database", err)
	sql, err = schemaFS.ReadFile("schema.sql")
	util.CheckFatal("failed to read schema", err)
	err = sqlitex.ExecScript(result.Conn, string(sql))
	util.CheckFatal("error applying schema", err)

	sql, err = schemaFS.ReadFile("insert.sql")
	util.CheckFatal("failed to read embedded file", err)
	result.QInsertBug = string(sql)

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

	return result
}

func (db *DB) GetDistinctUsers(resultFunc func(stmt *sqlite.Stmt) error) {

	execOptions := &sqlitex.ExecOptions{
		ResultFunc: resultFunc,
	}

	err := sqlitex.ExecuteTransient(db.Conn, db.QDistinctUsers, execOptions)
	util.CheckFatal("Failed to execute query", err)
}

func (db *DB) InsertBug(bug bugzilla.Bug) error {
	bugJson, err := json.Marshal(bug)
	util.CheckFatal("error marshalling bug JSON", err)

	execOptions := sqlitex.ExecOptions{
		Args: []interface{}{
			bug.ID,
			bug.CreationTime,
			bug.Creator,
			string(bugJson),
		},
	}

	err = sqlitex.ExecuteTransient(db.Conn, db.QInsertBug, &execOptions)
	util.CheckFatal("error executing insert bug statement", err)
	return nil
}
