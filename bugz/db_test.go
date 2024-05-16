package bugz

import (
	"testing"
	"zombiezen.com/go/sqlite/sqlitex"
	)

func TestCreateAndInitializeDatabase(t *testing.T) {
	db, err := CreateAndInitializeDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create and initialize database: %v", err)
	}
	defer db.Close()

	//var execOptions sqlitex.ExecOptions

	// Insert sample data into the bugs table
	insertText := `INSERT INTO bugs (id, CreationTime, Creator, Summary, OtherFieldsJSON) VALUES (1, '2077-10-23 09:42:00', 'John Dead', 'Sample Bug', '{}')`

	err := sqlitex.ExecuteTransient(db, insertText, nil)

	if err != nil {
    	    t.Fatalf("Error executing insert statement: %v", err)
	}

	// Prepare the statement for querying the bugs table
	prepareText := `SELECT id, CreationTime, Creator, Summary, OtherFieldsJSON FROM bugs WHERE id = 1`
	
	stmt, err := db.Prepare(prepareText)
	
	if err != nil {
		t.Fatalf("Failed to prepare select statement: %v", err)
	}
	defer stmt.Finalize()

	// Execute the query and retrieve the results
	hasRow, err := stmt.Step()
	if err != nil {
		t.Fatalf("Error stepping through rows: %v", err)
	}
	if !hasRow {
		t.Fatalf("No rows found for id = 1")
	}

	id := stmt.ColumnInt64(0)
	creationTime := stmt.ColumnText(1)
	creator := stmt.ColumnText(2)
	summary := stmt.ColumnText(3)
	otherFieldsJSON := stmt.ColumnText(4)

	// Verify the retrieved data matches the sample data
	if id != 1 || creationTime != "2077-10-23 09:42:00" || creator != "John Dead" || summary != "Sample Bug" || otherFieldsJSON != "{}" {
		t.Fatalf("Retrieved data does not match sample data")
	}
}
