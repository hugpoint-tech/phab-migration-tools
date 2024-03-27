// vim: set expandtab ts=4 sw=4 sts=4 :
package bugz

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"net/http"
	"net/url"
	"os"
	. "hugpoint.tech/freebsd/forge/util"
)

type BugzClient struct {
	token string
	URL   string

	http *http.Client
}

type BugzLoginResponse struct {
	Id    int    `json:"id"`
	Token string `json:"token"`
}

func NewBugzClient() *BugzClient {
	login := os.Getenv("BUGZILLA_LOGIN") //Retrieve env var values and check if they are empty
	password := os.Getenv("BUGZILLA_PASSWORD")
	if login == "" || password == "" {
		panic("BUGZILLA_LOGIN or BUGZILLA_PASSWORD is not set")
	}

	bugz := &BugzClient{
		URL:   "https://bugs.freebsd.org/bugzilla/rest",
		token: "",
		http:  &http.Client{},
	}

	/*formData := url.Values{}
	formData.Set("login", login)
	formData.Set("password", password)

	response, _ := bugz.http.Get(bugz.URL + "/login?" + formData.Encode())
	//CheckFatal("NewBugzClient bugz.http.Get", err)
	defer response.Body.Close()

	/*if response.StatusCode != http.StatusOK {
		CheckFatal("NewBugzClient", fmt.Errorf("login failed, status code: %d", response.StatusCode))
	}

	var loginResponse BugzLoginResponse
	if err := json.NewDecoder(response.Body).Decode(&loginResponse); err != nil {
		CheckFatal("NewBugzClient", fmt.Errorf("error reading bugzilla login response body: %w", err))
	}

	if loginResponse.Token == "" {
		CheckFatal("NewBugzClient loginResponse.Token", fmt.Errorf("login token is empty"))
	}
	bugz.Token = loginResponse.Token*/

	return bugz
}

func (bc *BugzClient) GetToken() (string, error) {
	return GetToken()
}

// DownloadAllBugs downloads all bugs from the Bugzilla API and saves them to individual JSON files.
func (bc *BugzClient) DownloadAllBugs() error {
	apiURL := "https://bugs.freebsd.org/bugzilla/rest/bug"

	// Create a 'bugs' folder if it doesn't exist
	err := os.MkdirAll("bugs", os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("Error creating 'users' folder: %v\n", err)
	}

	// Specify the pagination parameters
	pageSize := 1000
	pageNumber := 0

	for {
		// Create query parameters
		params := url.Values{}
		params.Set("api_key", "val")
		params.Set("limit", fmt.Sprint(pageSize))
		params.Set("offset", fmt.Sprint((pageNumber)*pageSize))

		// Construct the full URL with query parameters
		fullURL := apiURL + "?" + params.Encode()

		// Make a GET request to the API
		response, err := http.Get(fullURL)
		if err != nil {
			return fmt.Errorf("Error making GET request to %s: %v\n", fullURL, err)
		}

		// Read the response body
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("Error reading response body from %s: %v\n", fullURL, err)
		}
		response.Body.Close()

		// Process the JSON data
		var bugsResponse map[string][]Bug
		err = json.Unmarshal(body, &bugsResponse)
		if err != nil {
			return fmt.Errorf("Error decoding JSON: %v\n", err)
		}

		// Iterate over bugs and write to individual files
		for _, bug := range bugsResponse["bugs"] {
			filename := filepath.Join("bugsJson", fmt.Sprintf("bug_%d.json", bug.ID))
			err := writeToFile(filename, bug)
			if err != nil {
				return fmt.Errorf("Error writing to file %s: %v\n", filename, err)
			}
		}

		// Check if there are more pages
		if len(bugsResponse["bugs"]) < pageSize {
			break
		}

		// Move to the next page
		pageNumber++
	}

	fmt.Println("Bug data written to individual files.")
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

func (bc *BugzClient) showBugs() {
	fmt.Println("showing bugs")
}

func (bc *BugzClient) listBugs() {
	fmt.Println("listing bugs")
}

func (bc *BugzClient) getIDs() (map[int]User, error) {
	return GetIDS()
}
