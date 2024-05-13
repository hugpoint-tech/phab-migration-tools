package bugz

import (
	"log"
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
		log.Fatalf("Error creating table: %v", err)
	}

	// Insert sample data into the bugs table
	if err := sqlitex.Exec(db, "INSERT INTO bugs (id, CreationTime, Creator, Summary, OtherFieldsJSON) VALUES (?, ?, ?, ?, ?)", nil, 1, "2077-10-23 09:42:00", "John Dead", "Sample Bug", `{"key": "value"}`); err != nil {
		return fmt.Errorf("error executing statement: %v", err)
	}

	// Query the bugs table to verify that the sample data was inserted correctly
	var (
		id             int
		creationTime   string
		creator        string
		summary        string
		otherFieldsJSON string
	)
	row := db.QueryRow("SELECT id, CreationTime, Creator, Summary, OtherFieldsJSON FROM bugs WHERE id = ?", 1)
	if err := row.Scan(&id, &creationTime, &creator, &summary, &otherFieldsJSON); err != nil {
		t.Fatalf("Failed to scan row: %v", err)
	}

	// Verify the retrieved data matches the sample data
	if id != 1 || creationTime != "2077-10-23 09:42:00" || creator != "John Dead" || summary != "Sample Bug" || otherFieldsJSON != `{"key": "value"}` {
		t.Fatalf("Retrieved data does not match sample data")
	}
}
