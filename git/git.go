package git

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"code.gitea.io/sdk/gitea"

	"hugpoint.tech/freebsd/forge/bugz"
)

type Client struct {
	gc *gitea.Client
}

func NewClient(host string) (*Client, error) {
	giteaLogin, ok := os.LookupEnv("GITEA_LOGIN")
	if !ok {
		return nil, fmt.Errorf("GITEA_LOGIN is not set")
	}
	giteaPassword, ok := os.LookupEnv("GITEA_PASSWORD")
	if !ok {
		return nil, fmt.Errorf("GITEA_PASSWORD is not set")
	}

	gc, err := gitea.NewClient(host, gitea.SetBasicAuth(giteaLogin, giteaPassword))
	cl := Client{
		gc: gc,
	}

	return &cl, err
}

func (c *Client) CreateBugz(bugsChan <-chan bugz.Bug) error {
	for bug := range bugsChan {
		fullBody, err := json.Marshal(bug)
		if err != nil {
			return err
		}

		iss, resp, err := c.gc.CreateIssue("olex.podustov", "testrepo", gitea.CreateIssueOption{
			Title:     bug.Summary,
			Body:      string(fullBody),
			Ref:       bug.URL,
			Assignees: []string{bug.AssignedTo, bug.AssignedToDetail.Email},
			Deadline:  &time.Time{},
			Milestone: 0,
			Labels:    []int64{},
			Closed:    false,
		})
		if err != nil {
			return err
		}
		fmt.Println("iss", iss)
		fmt.Println("iss resp", resp)
	}

	return nil
}

func (c *Client) CreateUsers(users []bugz.User) error {
	visibility := gitea.VisibleTypePublic
	for _, user := range users {
		usr, resp, err := c.gc.AdminCreateUser(gitea.CreateUserOption{
			SourceID:           int64(user.ID),
			LoginName:          user.Name,
			Username:           user.RealName, // in bugzilla it mainly matches - although - response field name - The login name of the user. Note that in some situations this is different than their email.
			FullName:           user.RealName,
			Email:              user.Email,
			Password:           "", // of course we don't have password from bugzilla
			MustChangePassword: nil,
			SendNotify:         false,
			Visibility:         &visibility, // Public, Limited, Private
		})
		if err != nil {
			return fmt.Errorf("user %+v err %w", user, err)
		}
		fmt.Println("gitea create user", usr)
		fmt.Println("gitea create user resp", resp)
	}

	return nil
}
