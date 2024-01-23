// vim: set expandtab ts=4 sw=4 sts=4 :
package bugz

/// https://bugzilla.readthedocs.io/en/latest/api/index.html

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	. "hugpoint.tech/freebsd/forge/util"
)

func (b *BugzClient) Bugs() *BugsAPI {
	result := &BugsAPI{
		client: b,
		params: make(url.Values),
	}

	result.params.Set("token", b.token)
	return result
}

type BugsAPI struct {
	client *BugzClient
	params url.Values
}

func (b *BugsAPI) GetAll() []Bug {
	batchSize := 1000
	b.params.Set("limit", strconv.Itoa(batchSize))

	offset := 0
	result := []Bug{}

	for {
		b.params.Set("offset", strconv.Itoa(offset))
		bugsResponse := &BugsResponse{}
		response, err := b.client.http.Get(b.client.url + "/bug?" + b.params.Encode())
		CheckFatal(err)

		err = json.NewDecoder(response.Body).Decode(&bugsResponse)
		CheckFatal(err)

		result = append(result, bugsResponse.Bugs...)
		fmt.Printf("\rReading bugzilla bugs: %d", len(result))
		if len(bugsResponse.Bugs) < batchSize {
			break
		} else {
			offset = offset + batchSize
		}
		response.Body.Close()
	}

	return result
}
