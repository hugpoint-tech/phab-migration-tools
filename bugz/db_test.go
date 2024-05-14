package bugz

import (
	"testing"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
	"fmt"
)

func TestCreateAndInitializeDatabase(t *testing.T) {
	db, err := CreateAndInitializeDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create and initialize database: %v", err)
	}
	defer db.Close()

	// Insert sample data into the bugs table
	if err := sqlitex.ExecuteTransient(db, `INSERT INTO bugs (id, CreationTime, Creator, Summary, OtherFieldsJSON) VALUES (1, '2077-10-23 09:42:00', 'John Dead', 'Sample Bug', '{}')`, nil); err != nil {
		t.Fatalf("Error executing insert statement: %v", err)
	}

	// Query the bugs table to verify that the sample data was inserted correctly
	err = sqlitex.ExecuteTransient(db, `SELECT id, CreationTime, Creator, Summary, OtherFieldsJSON FROM bugs WHERE id = 1`, nil{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			var (
				id             int64
				creationTime   string
				creator        string
				summary        string
				otherFieldsJSON string
			)

			hasRow, err := stmt.Step()
			if err != nil {
				return fmt.Errorf("Error stepping through rows: %v", err)
			}
			if !hasRow {
				return fmt.Errorf("No rows found for id = 1")
			}

			id = stmt.ColumnInt64(0)
			creationTime = stmt.ColumnText(1)
			creator = stmt.ColumnText(2)
			summary = stmt.ColumnText(3)
			otherFieldsJSON = stmt.ColumnText(4)

			// Verify the retrieved data matches the sample data
			if id != 1 || creationTime != "2077-10-23 09:42:00" || creator != "John Dead" || summary != "Sample Bug" || otherFieldsJSON != "{}" {
				t.Fatalf("Retrieved data does not match sample data")
			}

			return nil
		},
	})
	if err != nil {
		t.Fatalf("Error executing select statement: %v", err)
	}
}
