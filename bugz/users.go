package bugz

import (
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

func (u *UserAPI) Get(chanDone chan struct{}, chanUsers chan User) error {
	// offset := 0
	// batchSize := 50

	iterator := 0
	for {
		// u.params.Set("offset", strconv.Itoa(offset))
		// u.params.Set("limit", strconv.Itoa(batchSize))
		// u.params.Set("match", "*")
		u.params.Set("ids", strconv.Itoa(iterator))

		url := u.client.url + "/user?" + u.params.Encode()
		// fmt.Println("Get url", url)
		response, err := u.client.http.Get(url)
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

		fmt.Printf("\rReading bugzilla users begin, current iterator: %d %d", iterator, len(result.Users))
		fmt.Println(result)

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
