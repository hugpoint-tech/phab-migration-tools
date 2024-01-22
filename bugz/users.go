package bugz

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"

	. "hugpoint.tech/freebsd/forge/util"
)

type UserAPI struct {
	client *BugzClient
	params url.Values
}

func (self *BugzClient) UserAPI() *UserAPI {
	result := &UserAPI{
		client: self,
		params: make(url.Values),
	}

	result.params.Set("token", self.token)
	return result
}

func (self *UserAPI) Get() map[string]interface{} {
	result := make(map[string]interface{})
	self.params.Set("match", "*")

	response, err := self.client.http.Get(self.client.url + "/bug?" + self.params.Encode())
	CheckFatal(err)

	body, err := io.ReadAll(response.Body)
	CheckFatal(err)

	err = json.Unmarshal(body, &result)
	CheckFatal(err)

	fmt.Println(result)
	return result
}
