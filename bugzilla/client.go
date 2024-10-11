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
	Token string
	URL   string
	Http  *http.Client
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
		Token: "",
		Http:  &http.Client{},
	}

	formData := url.Values{}
	formData.Set("login", login)
	formData.Set("password", password)

	response, err := bc.Http.Get(bc.URL + "/login?" + formData.Encode())
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
	bc.Token = loginResponse.Token

	return bc
}

var (
	totalBugs      int        // Total bugs counter
	mu             sync.Mutex // Mutex for synchronizing access to the totalBugsDownloaded slice
	pageSize       = 1000
	goroutineCount = 20
)

// Define the downloadBugs function that each goroutine will run
func (bc *Client) downloadBugs(routineNumber int, wg *sync.WaitGroup, bugChan chan Bug, errChan chan error) {
	defer wg.Done() // Signal that this goroutine is done when function returns

	// Local counter to track bugs for this goroutine
	localDownloaded := 0
	pageNumber := routineNumber // Start each goroutine from its own page number
	apiURL := bc.URL + "/bug"   // Base API URL

	for {
		// Create query parameters for the API request
		params := url.Values{}
		params.Set("token", bc.Token)
		params.Set("limit", fmt.Sprint(pageSize))
		params.Set("offset", fmt.Sprint(pageNumber*pageSize)) // Calculate offset based on page number

		// Construct the full URL with query parameters
		fullURL := apiURL + "?" + params.Encode()

		// Make a GET request to the API
		response, err := bc.Http.Get(fullURL)
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

		// Send each bug individually through the bug channel
		for _, bug := range bugsResponse["bugs"] {
			bugChan <- bug
			localDownloaded++                                                          // Increment local counter for this goroutine
			fmt.Printf("Goroutine %d downloading bug ID: %d\n", routineNumber, bug.ID) // Print which bug is being downloaded
		}

		// Use a mutex to safely update the total number of bugs downloaded across all goroutines
		mu.Lock()
		totalBugs += localDownloaded
		mu.Unlock()

		// Stop if fewer bugs than pageSize are returned
		if len(bugsResponse["bugs"]) < pageSize {
			break
		}

		// Move to the next page for this goroutine
		pageNumber += goroutineCount
	}
	fmt.Printf("Total bugs: %d\n", totalBugs)
}

// DownloadBugzillaBugs function to start the download with goroutineCount goroutines
func (bc *Client) DownloadBugzillaBugs(ctx context.Context) (<-chan Bug, <-chan error) {
	// Channels for bugs and errors
	bugChan := make(chan Bug)
	errChan := make(chan error)

	// WaitGroup to track completion of all goroutines
	var wg sync.WaitGroup

	// Start goroutineCount goroutines
	for i := 0; i < goroutineCount; i++ {
		wg.Add(1)                                    // Increment the WaitGroup counter
		go bc.downloadBugs(i, &wg, bugChan, errChan) // i is the routineNumber, used to calculate the start page
	}

	// Close the channels once all goroutines are done
	go func() {
		wg.Wait() // Wait for all goroutines to complete
		close(bugChan)
		close(errChan)
	}()

	// Return the channels so the caller can read from them
	return bugChan, errChan
}

func (bc *Client) DownloadBugComments(bugID int) ([]Comment, error) {
	apiURL := fmt.Sprintf("%s/bug/%d/comment", bc.URL, bugID)
	params := url.Values{}
	params.Set("token", bc.Token)
	fullURL := apiURL + "?" + params.Encode()
	response, err := bc.Http.Get(fullURL)

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
	params.Set("token", bc.Token)

	fullURL := apiURL + "?" + params.Encode()
	response, err := bc.Http.Get(fullURL)
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
