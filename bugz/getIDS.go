import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func getIDS() (map[int]User, error) {
	// Specify the path to the directory containing JSON files
	directoryPath := "bugz/"

	fmt.Println("Get IDS from bugs")

	// Get a list of all JSON files in the directory
	files, err := os.ReadDir(directoryPath)
	if err != nil {
		fmt.Printf("Error reading directory: %v\n", err)
		return nil, nil
	}

	// Map to store extracted IDs and corresponding User objects
	idUserMap := make(map[int]User)

	// Iterate over each file in the directory
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			// Construct the full path to the JSON file
			filePath := filepath.Join(directoryPath, file.Name())

			// Read the JSON file
			jsonData, err := os.ReadFile(filePath)
			if err != nil {
				fmt.Printf("Error reading JSON file %s: %v\n", filePath, err)
				continue
			}

			// Process the JSON data
			var bug Bug
			err = json.Unmarshal(jsonData, &bug)
			if err != nil {
				fmt.Printf("Error decoding JSON from file %s: %v\n", filePath, err)
				continue
			}

			// Add the extracted IDs to the map
			ids := extractIDs(bug)
			for id, user := range ids {
				idUserMap[id] = user
			}
		}
	}

	return idUserMap, nil
}

func extractIDs(bug Bug) map[int]User {
	// Create a map to store the unique IDs and corresponding Users
	idUserMap := make(map[int]User)

	// Extracting 'id' fields
	addUserToMap(idUserMap, bug.AssignedToDetail)

	// Extract 'id' field from CCDetail
	for _, ccUser := range bug.CCDetail {
		addUserToMap(idUserMap, ccUser)
	}

	// Extract 'id' field from CreatorDetail
	addUserToMap(idUserMap, bug.CreatorDetail)

	return idUserMap
}

func addUserToMap(idUserMap map[int]User, user User) {
	// Check if the ID is not present in the map before adding
	if _, ok := idUserMap[user.ID]; !ok {
		idUserMap[user.ID] = user
	}
}
