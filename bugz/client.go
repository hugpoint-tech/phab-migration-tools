package bugz

import (
	"encoding/json"
	"fmt"
	. "hugpoint.tech/freebsd/forge/common/bugzilla"
	"hugpoint.tech/freebsd/forge/util"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

type BugzClient struct {
	token string
	URL   string
	http  *http.Client
}

type BugzLoginResponse struct {
	Id    int    `json:"id"`
	Token string `json:"token"`
}

func NewBugzClient() *BugzClient {
	login := os.Getenv("BUGZILLA_LOGIN") //Retrieve env var values and check if they are empty
	password := os.Getenv("BUGZILLA_PASSWORD")
	if login == "" || password == "" {
		log.Fatal("BUGZILLA_LOGIN or BUGZILLA_PASSWORD is not set")
	}

	bc := &BugzClient{
		URL:   "https://bugs.freebsd.org/bugzilla/rest",
		token: "",
		http:  &http.Client{},
	}

	formData := url.Values{}
	formData.Set("login", login)
	formData.Set("password", password)

	response, err := bc.http.Get(bc.URL + "/login?" + formData.Encode())
	util.CheckFatal("login and/or password incorrect", err)

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		util.Fatalf("login failed, status code: %d", response.StatusCode)
	}

	var loginResponse BugzLoginResponse
	err = json.NewDecoder(response.Body).Decode(&loginResponse)
	util.CheckFatal("error reading bugzilla login response body", err)

	if loginResponse.Token == "" {
		util.Fatal("login token is empty")
	}
	bc.token = loginResponse.Token

	return bc
}

// DownloadBugzillaBugs downloads all bugs from the Bugzilla API and saves them to individual JSON files.
func (bc *BugzClient) DownloadBugzillaBugs() ([]Bug, error) { // Make URL to bugs
	apiURL := bc.URL + "/bug"

	// Specify the pagination parameters
	pageSize := 1000
	pageNumber := 0

	// TODO: this is suboptimal - memory consumption will go through the roof.
	// We need to refactor this function to work like a resumable iterator.
	bugs := make([]Bug, 0, 200000)

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
			return nil, fmt.Errorf("error making GET request to %s: %v", fullURL, err)
		}

		// Read the response body
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body from %s: %v", fullURL, err)
		}
		response.Body.Close()

		// Process the JSON data
		var bugsResponse map[string][]Bug
		err = json.Unmarshal(body, &bugsResponse)
		if err != nil {
			return nil, fmt.Errorf("error decoding JSON: %v", err)
		}

		bugs = append(bugs, bugsResponse["bugs"]...)

		// Check if there are more pages
		if len(bugsResponse["bugs"]) < pageSize {
			break
		}

		// Move to the next page
		pageNumber++
	}

	return bugs, nil
}

func (bc *BugzClient) ShowBugs() error {
	fmt.Println("showing bugs")
	return nil
}

func (bc *BugzClient) ListBugs() error {
	fmt.Println("listing bugs")
	return nil
}

func (bc *BugzClient) DownloadBugComments(bugID int64) ([]Comment, error) {
	apiURL := fmt.Sprintf("%s/bug/%d/comment", bc.URL, bugID)
	params := url.Values{}
	params.Set("token", bc.token)
	fullURL := apiURL + "?" + params.Encode()
	response, err := bc.http.Get(fullURL)

	if err != nil {
		return nil, fmt.Errorf("error making GET request to %s: %v", fullURL, err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body from %s: %v", fullURL, err)
	}

	var commentsResponse CommentsResponse
	if err := json.Unmarshal(body, &commentsResponse); err != nil {
		return nil, fmt.Errorf("error decoding JSON: %v", err)
	}

	return commentsResponse.Bugs[int(bugID)].Comments, nil
}

func (bc *BugzClient) DownloadBugAttachments(bugID int64) ([]Attachment, error) {
	apiURL := fmt.Sprintf("%s/bug/%d/attachment", bc.URL, bugID)
	params := url.Values{}
	params.Set("token", bc.token)

	fullURL := apiURL + "?" + params.Encode()
	response, err := bc.http.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("error making GET request to %s: %v", fullURL, err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body from %s: %v", fullURL, err)
	}

	var attachmentsResponse AttachmentsResponse
	if err := json.Unmarshal(body, &attachmentsResponse); err != nil {
		return nil, fmt.Errorf("error decoding JSON: %v", err)
	}

	return attachmentsResponse.Bugs[int(bugID)], nil
}
