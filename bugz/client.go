// vim: set expandtab ts=4 sw=4 sts=4 :
package bugz

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	. "hugpoint.tech/freebsd/forge/util"
)

type BugzClient struct {
	Token string
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
		Token: "",
		http:  &http.Client{},
	}

	formData := url.Values{}
	formData.Set("login", login)
	formData.Set("password", password)

	response, err := bugz.http.Get(bugz.URL + "/login?" + formData.Encode())
	CheckFatal("NewBugzClient bugz.http.Get", err)
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		CheckFatal("NewBugzClient", fmt.Errorf("login failed, status code: %d", response.StatusCode))
	}

	var loginResponse BugzLoginResponse
	if err := json.NewDecoder(response.Body).Decode(&loginResponse); err != nil {
		CheckFatal("NewBugzClient", fmt.Errorf("error reading bugzilla login response body: %w", err))
	}

	if loginResponse.Token == "" {
		CheckFatal("NewBugzClient loginResponse.Token", fmt.Errorf("login token is empty"))
	}
	bugz.Token = loginResponse.Token

	return bugz
}

func (bc *BugzClient) GetToken() string {
	return bc.Token
}

// DownloadAllBugs downloads all bugs from the Bugzilla API and saves them to individual JSON files.
func (bc *BugzClient) DownloadAllBugs() error {
	return DownloadBugzillaBugs()
}
