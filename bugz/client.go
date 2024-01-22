// vim: set expandtab ts=4 sw=4 sts=4 :
package bugz

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

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
	// login, ok := os.LookupEnv("BUGZILLA_LOGIN")
	// if !ok {
	// 	panic("BUGZILLA_LOGIN is not set")
	// }
	// password, ok := os.LookupEnv("BUGZILLA_PASSWORD")
	// if !ok {
	// 	panic("BUGZILLA_PASSWORD is not set")
	// }
	login := ""
	password := ``
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
	CheckFatal(err)
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		CheckFatal(fmt.Errorf("login failed, status code: %d", response.StatusCode))
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("error reading bugzilla login response body: %w", err)
		CheckFatal(err)
	}

	err = json.Unmarshal(body, &loginResponse)
	CheckFatal(err)

	if loginResponse.Token == "" {
		err = fmt.Errorf("login token is empty")
		CheckFatal(err)
	}
	bugz.token = loginResponse.Token

	return bugz
}
