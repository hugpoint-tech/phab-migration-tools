package phab

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	. "hugpoint.tech/freebsd/forge/util"
)

type UserSearchAPI struct {
	client *PhabClient
	method string
	params url.Values
}

func (self *PhabClient) UserSearch() *UserSearchAPI {
	result := &UserSearchAPI{
		method: "user.search",
		client: self,
		params: make(url.Values),
	}

	result.params.Set("api.token", self.token)
	return result
}

func (self *UserSearchAPI) Get() []User {

	var response *UserResponse
	result := []User{}

	self.params.Set("order", "oldest")
	self.params.Set("limit", "100")

	for {
		response = &UserResponse{}
		http_response, err := http.PostForm(self.client.url+self.method, self.params)
		CheckFatal(err)

		body, err := ioutil.ReadAll(http_response.Body)
		CheckFatal(err)

		err = json.Unmarshal(body, response)
		CheckFatal(err)

		if response.ErrorCode != "" {
			panic(response.ErrorInfo)
		}

		result = append(result, response.Result.Users...)
		http_response.Body.Close()

		fmt.Printf("\rReading phab users: %d", len(result))

		if response.Result.Cursor.After != "" {
			self.params.Set("after", response.Result.Cursor.After)
		} else {
			self.params.Del("after")
			break
		}
	}
	fmt.Println()
	return result
}
