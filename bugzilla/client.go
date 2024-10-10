package bugzilla

import (
	"context"
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

func (bc *Client) DownloadBugzillaBugs(ctx context.Context) (<-chan Bug, <-chan error) {
	apiURL := bc.URL + "/bug"

	// Create channels for streaming bugs and errors
	bugChan := make(chan Bug)
	errChan := make(chan error)

	// Create a WaitGroup to manage goroutines
	var wg sync.WaitGroup

	// Pagination parameters
	pageSize := 1000
	totalBugs := 0 // Counter to track total bugs downloaded

	// Define the downloadBugs function
	var downloadBugs func(pageNumber int)
	downloadBugs = func(pageNumber int) {
		defer wg.Done()

		// Create query parameters
		params := url.Values{}
		params.Set("token", bc.token)
		params.Set("limit", fmt.Sprint(pageSize))
		params.Set("offset", fmt.Sprint(pageNumber*pageSize))

		// Construct the full URL with query parameters
		fullURL := apiURL + "?" + params.Encode()

		// Make a GET request to the API
		response, err := bc.http.Get(fullURL)
		if err != nil {
			errChan <- fmt.Errorf("error making GET request to %s: %v", fullURL, err)
			return
		}

		// Read the response body
		body, err := io.ReadAll(response.Body)
		if err != nil {
			errChan <- fmt.Errorf("error reading response body from %s: %v", fullURL, err)
			response.Body.Close()
			return
		}
		response.Body.Close()

		// Process the JSON data
		var bugsResponse map[string][]Bug
		err = json.Unmarshal(body, &bugsResponse)
		if err != nil {
			errChan <- fmt.Errorf("error decoding JSON: %v", err)
			return
		}

		// Track the number of bugs downloaded for the current page
		bugsDownloaded := len(bugsResponse["bugs"])

		// Send each bug individually through the channel
		for _, bug := range bugsResponse["bugs"] {
			select {
			case <-ctx.Done(): // Handle context cancellation
				errChan <- ctx.Err()
				return
			case bugChan <- bug:
				totalBugs++ // Increment total bugs counter for each bug
			}
		}

		// Print the number of bugs downloaded after each page
		fmt.Printf("Total bugs downloaded: %d\n", totalBugs)

		// If we received fewer bugs than pageSize, stop (no more pages)
		if bugsDownloaded < pageSize {
			return
		}

		// Launch the next page download in a new goroutine
		wg.Add(1)
		go downloadBugs(pageNumber + 1)
	}

	// Start downloading with the first page
	wg.Add(1)
	go downloadBugs(0)

	// Start a goroutine to close the channels after all downloads complete
	go func() {
		wg.Wait()
		close(bugChan)
		close(errChan)

		// Print final bug count
		fmt.Printf("Finished downloading. Total bugs downloaded: %d\n", totalBugs)
	}()

	return bugChan, errChan
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
