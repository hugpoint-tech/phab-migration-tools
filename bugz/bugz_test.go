package bugz

import "testing"

func TestExtractIDs(t *testing.T) {

	bug := Bug{
		AssignedToDetail: User{ID: 101},
		CCDetail:         []User{{ID: 102}, {ID: 103}},
		CreatorDetail:    User{ID: 104},
	}

	ids := extractIDs(bug)

	expectedIDs := map[int]User{
		101: {ID: 101},
		102: {ID: 102},
		103: {ID: 103},
		104: {ID: 104},
	}


	for id, user := range expectedIDs {
		if ids[id].ID != user.ID {
			t.Errorf("extractIDs returned incorrect ID for user %d", id)
		}
	}
}
