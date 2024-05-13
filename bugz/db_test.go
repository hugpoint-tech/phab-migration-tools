package bugz

import (
	"testing"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

func TestCreateAndInitializeDatabase(t *testing.T) {
	// Open the in-memory database connection
	db, err := sqlite.OpenConn(":memory:", sqlite.OpenReadWrite)
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}
	defer db.Close()

	// Execute script to create the bugs table
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS bugs (
		id INTEGER PRIMARY KEY,
		CreationTime TEXT,
		Creator TEXT,
		Summary TEXT,
		OtherFieldsJSON TEXT
	);`
	if err := sqlitex.ExecScript(db, createTableQuery); err != nil {
		t.Fatalf("Error creating table: %v", err)
	}

	// Insert sample data into the bugs table
	if err := sqlitex.Exec(db, "INSERT INTO bugs (id, CreationTime, Creator, Summary, OtherFieldsJSON) VALUES (?, ?, ?, ?, ?)", 1, "2077-10-23 09:42:00", "John Dead", "Sample Bug", `{"key": "value"}`); err != nil {
		t.Fatalf("error executing statement: %v", err)
	}

	// Query the bugs table to verify that the sample data was inserted correctly
stmt, err := db.Prepare("SELECT id, CreationTime, Creator, Summary, OtherFieldsJSON FROM bugs WHERE id = ?")
if err != nil {
    t.Fatalf("Failed to prepare statement: %v", err)
}
defer stmt.Finalize()

if hasRow, err := stmt.Step(); err != nil {
    t.Fatalf("Error retrieving data: %v", err)
} else if !hasRow {
    t.Fatalf("No rows found for id = 1")
}

var (
    id             int64
    creationTime   string
    creator        string
    summary        string
    otherFieldsJSON string
)
id = stmt.ColumnInt64(0)
creationTime = stmt.ColumnText(1)
creator = stmt.ColumnText(2)
summary = stmt.ColumnText(3)
otherFieldsJSON = stmt.ColumnText(4)

// Verify the retrieved data matches the sample data
if id != 1 || creationTime != "2077-10-23 09:42:00" || creator != "John Dead" || summary != "Sample Bug" || otherFieldsJSON != `{"key": "value"}` {
    t.Fatalf("Retrieved data does not match sample data")
}
}
