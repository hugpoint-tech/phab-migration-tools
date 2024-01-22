package bugz

import (
	"encoding/json"
	"fmt"
	"net/url"

	. "hugpoint.tech/freebsd/forge/util"
)

type UserAPI struct {
	client *BugzClient
	params url.Values
}

func (u *BugzClient) UserAPI() *UserAPI {
	result := &UserAPI{
		client: u,
		params: make(url.Values),
	}

	result.params.Set("token", u.token)
	return result
}

func (u *UserAPI) Get() map[string]interface{} {
	u.params.Set("match", "*")

	response, err := u.client.http.Get(u.client.url + "/bug?" + u.params.Encode())
	CheckFatal(err)

	result := make(map[string]interface{})
	err = json.NewDecoder(response.Body).Decode(&result)
	CheckFatal(err)

	fmt.Println(result)
	return result
}
