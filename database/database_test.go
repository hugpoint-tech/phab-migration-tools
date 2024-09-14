package database

import (
	"context"
	"fmt"
	"hugpoint.tech/freebsd/forge/common/bugzilla"
	"hugpoint.tech/freebsd/forge/util"
	"testing"
	"zombiezen.com/go/sqlite/sqlitex"
)

func TestCreateAndInitializeDatabase(t *testing.T) {
	db := New(":memory:")
	conn := db.pool.Get(context.Background())
	if conn != nil {
		util.Fatal("failed to open database connection")
	}
	defer conn.Close()

	// Define the execOptions for the insert query
	execOptions := sqlitex.ExecOptions{
		Args: []interface{}{1, "2077-10-23 09:42:00", "John Dead", "{}"},
	}

	insertQuery, err := schemaFS.ReadFile("insert.sql")
	if err != nil {
		t.Fatalf("Failed to read insert query: %v", err)
	}

	if err := sqlitex.ExecuteTransient(conn, string(insertQuery), &execOptions); err != nil {
		t.Fatalf("Error executing insert statement: %v", err)
	}

	// Read the select query from the embedded file
	selectQuery, err := schemaFS.ReadFile("select.sql")
	if err != nil {
		t.Fatalf("Failed to read select query: %v", err)
	}

	stmt, err := conn.Prepare(string(selectQuery))

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
	otherFieldsJSON := stmt.ColumnText(3)

	// Verify the retrieved data matches the sample data
	if id != 1 || creationTime != "2077-10-23 09:42:00" || creator != "John Dead" || otherFieldsJSON != "{}" {
		t.Fatalf("Retrieved data does not match sample data")
	}
}

func TestGetDistinctUsers(t *testing.T) {
	db := New(":memory:")
	conn := db.pool.Get(context.Background())
	if conn != nil {
		util.Fatal("failed to open database connection")
	}
	defer conn.Close()

	// Insert sample data
	sampleData := [][]interface{}{
		{1, "2077-10-23 09:42:00", "john.doe@example.com", `{"assigned_to": "john.doe@example.com", "assigned_to_detail": {"email": "john.doe@example.com", "name": "John Doe", "real_name": "John Doe"}}`},
		{2, "2077-10-23 10:00:00", "jane.alive@example.com", `{"creator_detail": {"email": "jane.alive@example.com", "name": "Jane Alive", "real_name": "Jane Alive"}}`},
		{3, "2077-10-23 01:42:00", "john.doe@example.com", `{"assigned_to": "john.doe@example.com", "assigned_to_detail": {"email": "john.doe@example.com", "name": "John Doe", "real_name": "John Doe"}}`},
	}

	insertQuery, err := schemaFS.ReadFile("insert.sql")
	if err != nil {
		t.Fatalf("Failed to read insert query: %v", err)
	}

	for _, args := range sampleData {
		if err := sqlitex.ExecuteTransient(conn, string(insertQuery), &sqlitex.ExecOptions{Args: args}); err != nil {
			t.Fatalf("Error executing insert statement: %v", err)
		}
	}

	users, err := db.GetDistinctUsers()
	if err != nil {
		t.Fatal(err)
	}

	expectedUsers := []bugzilla.User{
		{Email: "jane.alive@example.com", Name: "Jane Alive", RealName: "Jane Alive"},
		{Email: "john.doe@example.com", Name: "John Doe", RealName: "John Doe"},
	}

	// Create a map of expected users for quick lookup
	expectedUserMap := make(map[string]bugzilla.User)
	for _, u := range expectedUsers {
		expectedUserMap[u.Email] = u
	}

	// Function to compare two User structs
	compareUsers := func(a, b bugzilla.User) bool {
		return a.Email == b.Email && a.Name == b.Name && a.RealName == b.RealName
	}

	// Convert users to a map for easy comparison
	retrievedUserMap := make(map[string]bugzilla.User)
	for _, user := range users {
		retrievedUserMap[user.Email] = user
	}

	// Check if all expected users are present in the results
	for _, expectedUser := range expectedUsers {
		if retrievedUser, found := retrievedUserMap[expectedUser.Email]; !found || !compareUsers(retrievedUser, expectedUser) {
			t.Errorf("Missing or unexpected user: got %+v, expected %+v", retrievedUserMap[expectedUser.Email], expectedUser)
		}
	}

	// Check if there are no extra unexpected users
	for _, user := range users {
		if _, found := expectedUserMap[user.Email]; !found {
			t.Errorf("Unexpected user found: %+v", user)
		}
	}

	// Debug: Print all retrieved users
	for _, user := range users {
		fmt.Printf("Retrieved user: %+v\n", user)
	}
}
