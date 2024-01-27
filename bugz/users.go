package bugz

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
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

func (u *UserAPI) Get(ctx context.Context, startIndex int, chanDone chan struct{}, chanUsers chan User) error {
	statistics := make(map[int]int)
	defer func() { fmt.Println("response code statistics", statistics) }()

	for {
		u.params.Set("ids", strconv.Itoa(startIndex))
		response, err := u.client.http.Get(u.client.url + "/user?" + u.params.Encode())
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

		statistics[response.StatusCode]++
		fmt.Printf("\rReading bugzilla users begin, current startIndex: %d len %d status %d", startIndex, len(result.Users), response.StatusCode)

		for _, v := range result.Users {
			chanUsers <- v
		}

		if startIndex > math.MaxInt-10 {
			close(chanDone)
			break
		}

		startIndex++
	}

	return nil
}

func writeUsersToFiles(folder string, instances ...User) error {
	for _, data := range instances {
		fmt.Println("Inserting user into file:", data.ID)
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
