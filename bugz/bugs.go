// vim: set expandtab ts=4 sw=4 sts=4 :
package bugz

/// https://bugzilla.readthedocs.io/en/latest/api/index.html

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"sync"

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
	mu     sync.Mutex
	client *BugzClient
	params url.Values
}

func (b *BugsAPI) GetAll(chanDone chan struct{}, chanBug chan Bug, pullConcurrencyLevel int) {
	offset := 0
	limitator := make(chan int, pullConcurrencyLevel)

	for {
		select {
		case <-chanDone:
			return
		case limitator <- 1:
			go func(offset int) {
				b.getBugs(chanDone, chanBug, offset)
				<-limitator
			}(offset)
		}
		offset += 1000
	}
}

func (b *BugsAPI) getBugs(chanDone chan struct{}, chanBug chan Bug, offset int) {
	select {
	case <-chanDone:
		return
	default:
	}

	fmt.Printf("\rReading bugzilla bugs begin, current offset: %d", offset)

	batchSize := 1000
	b.mu.Lock()

	b.params.Set("limit", strconv.Itoa(batchSize))
	b.params.Set("offset", strconv.Itoa(offset))
	params := b.params.Encode()
	url := b.client.url
	b.mu.Unlock()

	bugzilla := NewBugzClient()
	response, err := bugzilla.http.Get(url + "/bug?" + params)
	CheckFatal(err)
	defer response.Body.Close()

	bugsResponse := &BugsResponse{}
	err = json.NewDecoder(response.Body).Decode(&bugsResponse)
	CheckFatal(err)

	for _, v := range bugsResponse.Bugs {
		chanBug <- v
	}

	fmt.Printf("\rReading bugzilla bugs done, current offset: %d", offset)
	if len(bugsResponse.Bugs) < batchSize {
		select {
		case <-chanDone:
			return
		default:
			close(chanDone)
		}
		return
	}
}
