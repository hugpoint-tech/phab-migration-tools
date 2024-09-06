package bugz

import (
	"encoding/json"
	"fmt"
	"hugpoint.tech/freebsd/forge/database"
	"hugpoint.tech/freebsd/forge/util"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

type BugzClient struct {
	token string
	URL   string
	http  *http.Client
	DB    *database.DB
}

type BugzLoginResponse struct {
	Id    int    `json:"id"`
	Token string `json:"token"`
}

func NewBugzClient(db *database.DB) (*BugzClient, error) {
	login := os.Getenv("BUGZILLA_LOGIN") //Retrieve env var values and check if they are empty
	password := os.Getenv("BUGZILLA_PASSWORD")
	if login == "" || password == "" {
		log.Fatal("BUGZILLA_LOGIN or BUGZILLA_PASSWORD is not set")
	}

	bc := &BugzClient{
		URL:   "https://bugs.freebsd.org/bugzilla/rest",
		token: "",
		http:  &http.Client{},
		DB:    db,
	}

	formData := url.Values{}
	formData.Set("login", login)
	formData.Set("password", password)

	response, err := bc.http.Get(bc.URL + "/login?" + formData.Encode())
	if err != nil {
		fmt.Printf("login and/or password incorrect")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		fmt.Printf("login failed, status code: %d", response.StatusCode)
	}

	var loginResponse BugzLoginResponse
	if err := json.NewDecoder(response.Body).Decode(&loginResponse); err != nil {
		fmt.Printf("error reading bugzilla login response body: %s", err)
	}

	if loginResponse.Token == "" {
		fmt.Printf("login token is empty")
	}
	bc.token = loginResponse.Token

	return bc, nil
}

func (bc *BugzClient) InsertBug(bug Bug) error {
	bugJSON, err := json.Marshal(bug)
	util.CheckFatal("error marshalling bug JSON", err)

	execOptions := sqlitex.ExecOptions{
		Args: []interface{}{
			bug.ID,
			bug.CreationTime,
			bug.Creator,
			string(bugJSON),
		},
	}

	err = sqlitex.ExecuteTransient(bc.DB.Conn, bc.DB.QInsertBug, &execOptions)
	util.CheckFatal("error executing insert bug statement", err)

	return nil
}

// DownloadBugzillaBugs downloads all bugs from the Bugzilla API and saves them to individual JSON files.
func (bc *BugzClient) DownloadBugzillaBugs() error { // Make URL to bugs
	apiURL := bc.URL + "/bug"

	// Create a 'bugs' folder if it doesn't exist
	err := os.MkdirAll("bugs", os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("error creating 'users' folder: %v", err)
	}

	// Specify the pagination parameters
	pageSize := 1000
	pageNumber := 0
	totalBugs := 0

	for {
		// Create query parameters
		params := url.Values{}
		params.Set("token", bc.token)
		params.Set("limit", fmt.Sprint(pageSize))
		params.Set("offset", fmt.Sprint((pageNumber)*pageSize))

		// Construct the full URL with query parameters
		fullURL := apiURL + "?" + params.Encode()

		// Make a GET request to the API
		response, err := bc.http.Get(fullURL)
		if err != nil {
			return fmt.Errorf("error making GET request to %s: %v", fullURL, err)
		}

		// Read the response body
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("error reading response body from %s: %v", fullURL, err)
		}
		response.Body.Close()

		// Process the JSON data
		var bugsResponse map[string][]Bug
		err = json.Unmarshal(body, &bugsResponse)
		if err != nil {
			return fmt.Errorf("error decoding JSON: %v", err)
		}

		for _, bug := range bugsResponse["bugs"] {
			if err := bc.InsertBug(bug); err != nil {
				return fmt.Errorf("error inserting bug %d: %v", bug.ID, err)
			}
		}

		// Update the total number of bugs downloaded
		totalBugs += len(bugsResponse["bugs"])

		// Print the number of bugs downloaded after each page
		fmt.Printf("Total bugs downloaded: %d\n", totalBugs)

		// Check if there are more pages
		if len(bugsResponse["bugs"]) < pageSize {
			break
		}

		// Move to the next page
		pageNumber++
	}

	return nil
}

func writeToFile(filename string, bug Bug) error {
	// Use json.MarshalIndent to preserve the original formatting
	indentedData, err := json.MarshalIndent(bug, "", "  ")
	if err != nil {
		return err
	}

	// Write the data to the specified file
	return os.WriteFile(filename, indentedData, 0644)
}

func (bc *BugzClient) ShowBugs() error {
	fmt.Println("showing bugs")
	return nil
}

func (bc *BugzClient) ListBugs() error {
	fmt.Println("listing bugs")
	return nil
}

func getIDS() (map[int]User, error) {
	// Specify the path to the directory containing JSON files
	directoryPath := "bugs"

	fmt.Println("Get IDS from bugs")

	// Get a list of all JSON files in the directory
	files, err := os.ReadDir(directoryPath)
	if err != nil {
		return nil, fmt.Errorf("error reading directory: %w", err)
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

	// Extract 'id' field from AssignedToDetail
	if _, ok := idUserMap[bug.AssignedToDetail.ID]; !ok {
		idUserMap[bug.AssignedToDetail.ID] = bug.AssignedToDetail
	}

	// Extract 'id' field from CCDetail
	for _, ccUser := range bug.CCDetail {
		if _, ok := idUserMap[ccUser.ID]; !ok {
			idUserMap[ccUser.ID] = ccUser
		}
	}

	// Extract 'id' field from CreatorDetail
	if _, ok := idUserMap[bug.CreatorDetail.ID]; !ok {
		idUserMap[bug.CreatorDetail.ID] = bug.CreatorDetail
	}

	return idUserMap
}

func (bc *BugzClient) DownloadBugzillaUsers() error {
	var users []User

	// Execute distinct query on the bugs table to retrieve unique user data
	bc.DB.GetDistinctUsers(func(stmt *sqlite.Stmt) error {
		user := User{
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
	})

	// Debug: Print all retrieved users
	for _, user := range users {
		fmt.Printf("Retrieved user: %+v\n", user)
	}

	userCount := 0
	// Insert each user into the users table
	for _, user := range users {
		execOptions := sqlitex.ExecOptions{
			Args: []interface{}{user.Email, user.Name, user.RealName},
		}

		if err := sqlitex.Execute(bc.DB.Conn, bc.DB.QInsertUsers, &execOptions); err != nil {
			return fmt.Errorf("error inserting user: %v", err)
		}
		userCount++
	}
	fmt.Printf("Downloaded %d unique users\n", userCount)
	return nil
}

func (bc *BugzClient) DownloadBugzillaComments(bugIDs []int64) error {
	var wg sync.WaitGroup
	taskChan := make(chan CommentTask, len(bugIDs)) // Channel to pass tasks between producer and consumer
	semaphore := make(chan struct{}, 20)            // Limit concurrent connections to 20

	// Start the consumers
	numConsumers := runtime.NumCPU() // Use number of CPU cores as consumer count
	for i := 0; i < numConsumers; i++ {
		wg.Add(1)
		go bc.commentConsumer(taskChan, &wg)
	}

	// Start the producer
	for _, bugID := range bugIDs {
		semaphore <- struct{}{} // Acquire slot for a connection
		go func(bugID int64) {
			defer func() { <-semaphore }() // Release slot after completion
			bc.commentProducer(bugID, taskChan)
		}(bugID)
	}

	// Wait for consumers to finish processing
	wg.Wait()
	close(taskChan)

	return nil
}

// Producer: Downloads comments for a specific bug and sends them to the channel
func (bc *BugzClient) commentProducer(bugID int64, taskChan chan<- CommentTask) {
	apiURL := fmt.Sprintf("%s/bug/%d/comment", bc.URL, bugID)
	params := url.Values{}
	params.Set("token", bc.token)

	fullURL := apiURL + "?" + params.Encode()
	maxRetries := 3
	var response *http.Response
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		response, err = bc.http.Get(fullURL)
		if err == nil {
			break
		}

		if err == io.EOF || os.IsTimeout(err) {
			log.Printf("Error making GET request to %s (attempt %d/%d): %v", fullURL, attempt, maxRetries, err)
			time.Sleep(time.Duration(attempt) * time.Second) // Exponential backoff
			continue
		} else {
			log.Printf("Fatal error making GET request to %s: %v", fullURL, err)
			return
		}
	}

	if err != nil {
		log.Printf("Failed to make GET request to %s after %d attempts: %v", fullURL, maxRetries, err)
		return
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("Error reading response body from %s: %v", fullURL, err)
		return
	}

	// Log the response if JSON decoding fails
	var commentsResponse CommentsResponse
	if err := json.Unmarshal(body, &commentsResponse); err != nil {
		log.Printf("Error decoding JSON from %s: %v. Response body: %s", fullURL, err, string(body))
		return
	}

	comments := commentsResponse.Bugs[int(bugID)].Comments
	taskChan <- CommentTask{BugID: bugID, Comments: comments}
}

// Mutual exclusion lock variable
var dbMutex sync.Mutex

// Consumer: Reads from the channel and inserts comments into the database
func (bc *BugzClient) commentConsumer(taskChan <-chan CommentTask, wg *sync.WaitGroup) {
	defer wg.Done()

	db := database.New("bugsNew.db")

	insertQuery := db.QInsertComments
	if insertQuery == "" {
		log.Fatalf("Error: Insert comment query is empty")
	}

	for task := range taskChan {
		commentCount := 0
		for _, comment := range task.Comments {
			execOptions := sqlitex.ExecOptions{
				Args: []interface{}{
					comment.ID,
					comment.BugID,
					comment.AttachmentID,
					comment.CreationTime,
					comment.Creator,
					comment.Text,
				},
			}

			dbMutex.Lock() // Ensure only one goroutine is writing to the database at a time
			if err := sqlitex.Execute(db.Conn, insertQuery, &execOptions); err != nil {
				log.Printf("Error inserting comment for bug %d: %v", task.BugID, err)
				dbMutex.Unlock()
				continue
			}
			dbMutex.Unlock()

			commentCount++
		}
		fmt.Printf("Processed %d comments for bug %d\n", commentCount, task.BugID)
	}
}

func (bc *BugzClient) DownloadBugzillaAttachments(bugIDs []int64) error {
	var wg sync.WaitGroup
	taskChan := make(chan AttachmentTask, len(bugIDs)) // Channel to pass tasks between producer and consumer
	semaphore := make(chan struct{}, 20)               // Limit concurrent connections to 20

	// Start the consumers
	numConsumers := runtime.NumCPU() // Use number of CPU cores as consumer count
	for i := 0; i < numConsumers; i++ {
		wg.Add(1)
		go bc.attachmentConsumer(taskChan, &wg)
	}

	// Start the producer
	for _, bugID := range bugIDs {
		semaphore <- struct{}{} // Acquire slot for a connection
		go func(bugID int64) {
			defer func() { <-semaphore }() // Release slot after completion
			bc.attachmentProducer(bugID, taskChan)
		}(bugID)
	}

	// Wait for consumers to finish processing
	wg.Wait()
	close(taskChan)

	return nil
}

// Producer: Downloads attachments for a specific bug and sends them to the channel
func (bc *BugzClient) attachmentProducer(bugID int64, taskChan chan<- AttachmentTask) {
	apiURL := fmt.Sprintf("%s/bug/%d/attachment", bc.URL, bugID)
	params := url.Values{}
	params.Set("token", bc.token)

	fullURL := apiURL + "?" + params.Encode()
	maxRetries := 3
	var response *http.Response
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		response, err = bc.http.Get(fullURL)
		if err == nil {
			break
		}

		if err == io.EOF || os.IsTimeout(err) {
			log.Printf("Error making GET request to %s (attempt %d/%d): %v", fullURL, attempt, maxRetries, err)
			time.Sleep(time.Duration(attempt) * time.Second) // Exponential backoff
			continue
		} else {
			log.Printf("Fatal error making GET request to %s: %v", fullURL, err)
			return
		}
	}
	if err != nil {
		log.Printf("Failed to make GET request to %s after %d attempts: %v", fullURL, maxRetries, err)
		return
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("Error reading response body from %s: %v", fullURL, err)
		return
	}

	var attachmentsResponse AttachmentResponse
	if err := json.Unmarshal(body, &attachmentsResponse); err != nil {
		fmt.Errorf("error decoding JSON: %v", err)
	}

	attachments := attachmentsResponse.Bugs[fmt.Sprintf("%d", bugID)]
	taskChan <- AttachmentTask{BugID: bugID, Attachments: attachments}
}

// Consumer: Reads from the channel and inserts attachments into the database
func (bc *BugzClient) attachmentConsumer(taskChan <-chan AttachmentTask, wg *sync.WaitGroup) {
	defer wg.Done()

	db := database.New("bugsNew.db")

	insertQuery := db.QInsertAttachments
	if insertQuery == "" {
		log.Fatalf("Error: Insert attachment query is empty")
	}

	for task := range taskChan {
		attachmentsCount := 0
		for _, attachment := range task.Attachments {
			execOptions := sqlitex.ExecOptions{
				Args: []interface{}{
					attachment.ID,
					attachment.BugID,
					attachment.CreationTime,
					attachment.Creator,
					attachment.Summary,
					attachment.Data,
				},
			}
			dbMutex.Lock() // Ensure only one goroutine is writing to the database at a time
			if err := sqlitex.Execute(db.Conn, insertQuery, &execOptions); err != nil {
				log.Printf("Error inserting attachment for bug %d: %v", task.BugID, err)
				dbMutex.Unlock()
				continue
			}
			dbMutex.Unlock()

			attachmentsCount++
		}
		fmt.Printf("Processed %d attachments for bug %d\n", attachmentsCount, task.BugID)
	}
}

func FetchBugsFromDatabase(db *sqlite.Conn) ([]Bug, error) {
	var bugs []Bug
	query := "SELECT id FROM bugs"
	stmt := db.Prep(query)
	defer stmt.Finalize()

	for {
		hasNext, err := stmt.Step()
		if err != nil {
			return nil, fmt.Errorf("error fetching bugs: %v", err)
		}
		if !hasNext {
			break
		}
		bugID := stmt.ColumnInt64(0)
		bugs = append(bugs, Bug{ID: int(bugID)})
	}

	return bugs, nil
}
