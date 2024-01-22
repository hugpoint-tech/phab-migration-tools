// vim: set expandtab ts=4 sw=4 sts=4 :
package bugz

/// https://bugzilla.readthedocs.io/en/latest/api/index.html

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"

	. "hugpoint.tech/freebsd/forge/util"
)

func (self *BugzClient) Bugs() *BugsAPI {
	result := &BugsAPI{
		client: self,
		params: make(url.Values),
	}

	result.params.Set("token", self.token)
	return result
}

type BugsAPI struct {
	client *BugzClient
	params url.Values
}

func (self *BugsAPI) GetAll() []Bug {
	batch_size := 1000
	offset := 0
	result := []Bug{}
	var bugsResponse *BugsResponse
	self.params.Set("limit", strconv.Itoa(batch_size))

	for {
		self.params.Set("offset", strconv.Itoa(offset))
		bugsResponse = &BugsResponse{}
		response, err := self.client.http.Get(self.client.url + "/bug?" + self.params.Encode())
		CheckFatal(err)

		body, err := io.ReadAll(response.Body)
		CheckFatal(err)

		err = json.Unmarshal(body, &bugsResponse)
		CheckFatal(err)

		result = append(result, bugsResponse.Bugs...)
		fmt.Printf("\rReading bugzilla bugs: %d", len(result))
		if len(bugsResponse.Bugs) < batch_size {
			break
		} else {
			offset = offset + batch_size
		}
		response.Body.Close()
	}

	return result
}
