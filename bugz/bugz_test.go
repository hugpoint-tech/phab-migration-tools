package bugz

import "testing"

func TestExtractIDs(t *testing.T) {
	
	bug := Bug{
		AssignedToDetail: User{ID: 101},
		CCDetail:         []User{{ID: 102}, {ID: 103}},
		CreatorDetail:    User{ID: 104},
	}
	
	expectedIDs := map[int]User{
		101: {ID: 101},
		102: {ID: 102},
		103: {ID: 103},
		104: {ID: 104},
	}
	
	actualIDs := extractIDs(bug)
	
	if len(actualIDs) != len(expectedIDs) {
		t.Errorf("Number of actual IDs does not match number of expected IDs")
		return
	}

	for id, expectedUser := range expectedIDs {
		actualUser, ok := actualIDs[id]
		if !ok {
			t.Errorf("Expected ID %d not found in actual IDs", id)
			return
		}
		if actualUser.ID != expectedUser.ID {
			t.Errorf("Actual user ID %d does not match expected user ID %d for ID %d", actualUser.ID, expectedUser.ID, id)
		}
	}
}
