package bugzilla

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
	"sync"
)

type Client struct {
	token string
	URL   string
	http  *http.Client
}

type loginResponse struct {
	Id    int    `json:"id"`
	Token string `json:"token"`
}

func NewClient() Client {
	login := os.Getenv("BUGZILLA_LOGIN") //Retrieve env var values and check if they are empty
	password := os.Getenv("BUGZILLA_PASSWORD")
	if login == "" || password == "" {
		log.Fatal("BUGZILLA_LOGIN or BUGZILLA_PASSWORD is not set")
	}

	bc := Client{
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

	var loginResponse loginResponse
	err = json.NewDecoder(response.Body).Decode(&loginResponse)
	util.CheckFatal("error reading bugzilla login response body", err)

	if loginResponse.Token == "" {
		util.Fatal("login token is empty")
	}
	bc.token = loginResponse.Token

	return bc
}

var (
	totalBugs      int        // Total bugs counter
	mu             sync.Mutex // Mutex for synchronizing access to the totalBugs
	pageSize       = 1000
	goroutineCount = 20
)

func (bc *Client) DownloadBugs(offset, limit int) ([]Bug, error) {
	// Prepare the request URL with query parameters for pagination (offset and limit)
	reqURL := fmt.Sprintf("%s/bug?offset=%d&limit=%d", bc.URL, offset, limit)

	// Create a new request
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set the authorization token in the headers, if needed
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", bc.token))

	// Execute the request
	resp, err := bc.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check for a non-200 status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse the response body
	var result struct {
		Bugs []Bug `json:"bugs"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Return the list of bugs
	return result.Bugs, nil
}

func (bc *Client) DownloadBugComments(bugID int) ([]Comment, error) {
	apiURL := fmt.Sprintf("%s/bug/%d/comment", bc.URL, bugID)
	params := url.Values{}
	params.Set("token", bc.token)
	fullURL := apiURL + "?" + params.Encode()
	response, err := bc.http.Get(fullURL)

	if err != nil {
		return nil, fmt.Errorf("error making GET request to %s: %v", fullURL, err)
	}
	defer response.Body.Close()

	if response.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("comment request error: %s", response.Status)
	}

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

func (bc *Client) DownloadBugAttachments(bugID int) ([]Attachment, error) {
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
