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
	token string
	url   string

	http *http.Client
}

type BugzLoginResponse struct {
	Id    int    `json:"id"`
	Token string `json:"token"`
}

func NewBugzClient() BugzClient {
	login, ok := os.LookupEnv("BUGZILLA_LOGIN")
	if !ok {
		panic("BUGZILLA_LOGIN is not set")
	}
	password, ok := os.LookupEnv("BUGZILLA_PASSWORD")
	if !ok {
		panic("BUGZILLA_PASSWORD is not set")
	}
	var loginResponse BugzLoginResponse

	bugz := BugzClient{
		url:   "https://bugs.freebsd.org/bugzilla/rest",
		token: "",
		http:  &http.Client{
			// Timeout: time.Second * 10,
		},
	}
	formData := url.Values{}
	formData.Set("login", login)
	formData.Set("password", password)

	response, err := bugz.http.Get(bugz.url + "/login?" + formData.Encode())
	CheckFatal("NewBugzClient bugz.http.Get", err)
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		CheckFatal("NewBugzClient", fmt.Errorf("login failed, status code: %d", response.StatusCode))
	}

	if err := json.NewDecoder(response.Body).Decode(&loginResponse); err != nil {
		CheckFatal("NewBugzClient", fmt.Errorf("error reading bugzilla login response body: %w", err))
	}

	if loginResponse.Token == "" {
		err = fmt.Errorf("login token is empty")
		CheckFatal("NewBugzClient loginResponse.Token", err)
	}
	bugz.token = loginResponse.Token

	return bugz
}
