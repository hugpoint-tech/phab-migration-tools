package bugz

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
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

func (u *UserAPI) Get(ctx context.Context, chanDone chan struct{}, chanUsers chan User) error {
	iterator := 0
	for {
		u.params.Set("ids", strconv.Itoa(iterator))
		url := u.client.url + "/user?" + u.params.Encode()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		response, err := u.client.http.Do(req)
		if err != nil {
			return err
		}

		var result UsersResponse
		if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
			return err
		}
		if err := writeUsersToFiles("../bsdata/bugzilla-users", result.Users...); err != nil {
			return err
		}

		fmt.Printf("\rReading bugzilla users begin, current iterator: %d len %d status %d result %+v", iterator, len(result.Users), response.StatusCode, result)

		for _, v := range result.Users {
			chanUsers <- v
		}

		if iterator > math.MaxInt-10 {
			close(chanDone)
			break
		}

		iterator++
	}

	return nil
}

func writeUsersToFiles(folder string, instances ...User) error {
	for i, data := range instances {
		fmt.Printf("\rInserting user into file: %d", i)
		bytes, err := json.Marshal(data)
		if err != nil {
			return err
		}

		if err := os.WriteFile(fmt.Sprintf("%s/%d.json", folder, data.ID), bytes, 0o644); err != nil {
			return err
		}
	}
	return nil
}
