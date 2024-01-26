package bugz

import (
	"encoding/json"
	"fmt"
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

func (u *UserAPI) Get(chanUsers chan User) error {
	offset := 0
	batchSize := 1000
	offset += batchSize

	u.params.Set("match", "*")
	u.params.Set("limit", strconv.Itoa(batchSize))
	u.params.Set("offset", strconv.Itoa(offset))

	response, err := u.client.http.Get(u.client.url + "/user?" + u.params.Encode())
	if err != nil {
		return err
	}

	var result UsersResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return err
	}

	if err := writeUsersToFiles("bugzilla-users", result.Users...); err != nil {
		return err
	}

	// var buf bytes.Buffer
	// newReader := io.TeeReader(response.Body, &buf)
	// bytes, err := io.ReadAll(&buf)
	// if err != nil {
	// 	return err
	// }
	// fmt.Println("user data", string(bytes))

	fmt.Println("len(result.Users) - ", len(result.Users))
	fmt.Println(result)

	for _, v := range result.Users {
		chanUsers <- v
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

		if err := os.WriteFile(fmt.Sprintf("./%s/%d.json", folder, data.ID), bytes, 0o644); err != nil {
			return err
		}
	}
	return nil
}
