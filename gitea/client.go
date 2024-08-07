package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var jsonHeader = http.Header{"content-type": []string{"application/json"}}

// Version return the library version
func Version() string {
	return "0.16.0"
}

// Client represents a thread-safe Gitea API client.
type Client struct {
	url            string
	accessToken    string
	username       string
	password       string
	otp            string
	sudo           string
	userAgent      string
	debug          bool
	client         *http.Client
	ctx            context.Context
	mutex          sync.RWMutex
	getVersionOnce sync.Once
	ignoreVersion  bool // only set by SetGiteaVersion so don't need a mutex lock
}

// Response represents the Gitea response
type Response struct {
	*http.Response

	FirstPage int
	PrevPage  int
	NextPage  int
	LastPage  int
}

// ClientOption are functions used to init a new client
type ClientOption func(*Client) error

// SetToken sets the token for the client.
func SetToken() ClientOption {
	return func(c *Client) error {
		token := os.Getenv("GITEA_TOKEN")
		if token == "" {
			fmt.Fprintf(os.Stderr, "GITEA_TOKEN environment variable not set\n")
			os.Exit(1)
		}
		c.accessToken = token

		return nil
	}
}

// User represents a Gitea user
type User struct {
	ID               int64     `json:"id"`
	UserName         string    `json:"login"`
	LoginName        string    `json:"login_name"`
	SourceID         int64     `json:"source_id"`
	FullName         string    `json:"full_name"`
	Email            string    `json:"email"`
	AvatarURL        string    `json:"avatar_url"`
	Language         string    `json:"language"`
	IsAdmin          bool      `json:"is_admin"`
	LastLogin        time.Time `json:"last_login"`
	Created          time.Time `json:"created"`
	Restricted       bool      `json:"restricted"`
	IsActive         bool      `json:"active"`
	ProhibitLogin    bool      `json:"prohibit_login"`
	Location         string    `json:"location"`
	Website          string    `json:"website"`
	Description      string    `json:"description"`
	FollowerCount    int       `json:"followers_count"`
	FollowingCount   int       `json:"following_count"`
	StarredRepoCount int       `json:"starred_repos_count"`
}

// Permission represents repository permissions
type Permission struct {
	Admin bool `json:"admin"`
	Push  bool `json:"push"`
	Pull  bool `json:"pull"`
}

// InternalTracker represents internal issue tracker configuration
type InternalTracker struct {
}

// ExternalTracker represents external issue tracker configuration
type ExternalTracker struct {
}

// ExternalWiki represents external wiki configuration
type ExternalWiki struct {
}

// MergeStyle represents the merge style of the repository
type MergeStyle string

// Repository represents a Gitea repository
type Repository struct {
	ID                        int64            `json:"id"`
	Owner                     *User            `json:"owner"`
	Name                      string           `json:"name"`
	FullName                  string           `json:"full_name"`
	Description               string           `json:"description"`
	Empty                     bool             `json:"empty"`
	Private                   bool             `json:"private"`
	Fork                      bool             `json:"fork"`
	Template                  bool             `json:"template"`
	Parent                    *Repository      `json:"parent"`
	Mirror                    bool             `json:"mirror"`
	Size                      int              `json:"size"`
	HTMLURL                   string           `json:"html_url"`
	SSHURL                    string           `json:"ssh_url"`
	CloneURL                  string           `json:"clone_url"`
	OriginalURL               string           `json:"original_url"`
	Website                   string           `json:"website"`
	Stars                     int              `json:"stars_count"`
	Forks                     int              `json:"forks_count"`
	Watchers                  int              `json:"watchers_count"`
	OpenIssues                int              `json:"open_issues_count"`
	OpenPulls                 int              `json:"open_pr_counter"`
	Releases                  int              `json:"release_counter"`
	DefaultBranch             string           `json:"default_branch"`
	Archived                  bool             `json:"archived"`
	Created                   time.Time        `json:"created_at"`
	Updated                   time.Time        `json:"updated_at"`
	Permissions               *Permission      `json:"permissions,omitempty"`
	HasIssues                 bool             `json:"has_issues"`
	InternalTracker           *InternalTracker `json:"internal_tracker,omitempty"`
	ExternalTracker           *ExternalTracker `json:"external_tracker,omitempty"`
	HasWiki                   bool             `json:"has_wiki"`
	ExternalWiki              *ExternalWiki    `json:"external_wiki,omitempty"`
	HasPullRequests           bool             `json:"has_pull_requests"`
	HasProjects               bool             `json:"has_projects"`
	HasReleases               bool             `json:"has_releases,omitempty"`
	HasPackages               bool             `json:"has_packages,omitempty"`
	HasActions                bool             `json:"has_actions,omitempty"`
	IgnoreWhitespaceConflicts bool             `json:"ignore_whitespace_conflicts"`
	AllowMerge                bool             `json:"allow_merge_commits"`
	AllowRebase               bool             `json:"allow_rebase"`
	AllowRebaseMerge          bool             `json:"allow_rebase_explicit"`
	AllowSquash               bool             `json:"allow_squash_merge"`
	AvatarURL                 string           `json:"avatar_url"`
	Internal                  bool             `json:"internal"`
	MirrorInterval            string           `json:"mirror_interval"`
	MirrorUpdated             time.Time        `json:"mirror_updated,omitempty"`
	DefaultMergeStyle         MergeStyle       `json:"default_merge_style"`
}

// Organization represents a Gitea organization
type Organization struct {
	ID          int64  `json:"id"`
	UserName    string `json:"username"`
	FullName    string `json:"full_name"`
	AvatarURL   string `json:"avatar_url"`
	Description string `json:"description"`
	Website     string `json:"website"`
	Location    string `json:"location"`
	Visibility  string `json:"visibility"`
}

// ListOptions contains common options for listing resources
type ListOptions struct {
	// Setting Page to -1 disables pagination on endpoints that support it.
	// Page numbering starts at 1.
	Page int
	// The default value depends on the server config DEFAULT_PAGING_NUM
	// The highest valid value depends on the server config MAX_RESPONSE_ITEMS
	PageSize int
}

// ListReposOptions options for listing repositories
type ListReposOptions struct {
	ListOptions
}

// ListOrgsOptions options for listing organizations
type ListOrgsOptions struct {
	ListOptions
}

// NewClient creates a new Gitea client
func NewClient(url string, options ...ClientOption) (*Client, error) {
	client := &Client{
		url:    strings.TrimSuffix(url, "/"),
		client: &http.Client{},
		ctx:    context.Background(),
	}
	for _, opt := range options {
		if err := opt(client); err != nil {
			return nil, err
		}
	}

	return client, nil
}

// ListMyRepos lists the repositories of the authenticated user
func (c *Client) ListMyRepos(opt ListReposOptions) ([]*Repository, *Response, error) {
	url := fmt.Sprintf("%s/user/repos?page=%d&per_page=%d", c.url, opt.Page, opt.PageSize)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header = jsonHeader

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, errors.New("failed to fetch repositories")
	}

	var repos []*Repository
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, nil, err
	}

	return repos, &Response{Response: resp}, nil
}

// ListMyOrgs lists the organizations of the authenticated user
func (c *Client) ListMyOrgs(opt ListOrgsOptions) ([]*Organization, *Response, error) {
	url := fmt.Sprintf("%s/user/orgs?page=%d&per_page=%d", c.url, opt.Page, opt.PageSize)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header = jsonHeader

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, errors.New("failed to fetch organizations")
	}

	var orgs []*Organization
	if err := json.NewDecoder(resp.Body).Decode(&orgs); err != nil {
		return nil, nil, err
	}

	return orgs, &Response{Response: resp}, nil
}

// GetUser retrieves the user information.
func (c *Client) GetUser() (*User, error) {
	req, err := http.NewRequest("GET", c.url+"/user", nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "token "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error response from server: %s", resp.Status)
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &user, nil
}

func main() {
	client, err := NewClient("https://gitcvt.hugpoint.tech/api/v1", SetToken())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating client: %v\n", err)
		os.Exit(1)
	}

	user, err := client.GetUser()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching user: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("User: %s\n", user.UserName)
	fmt.Printf("Full Name: %s\n", user.FullName)
	fmt.Printf("Email: %s\n", user.Email)

	repos, _, err := client.ListMyRepos(ListReposOptions{ListOptions{Page: 1, PageSize: 10}})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing repos: %v\n", err)
	}

	for _, repo := range repos {
		fmt.Printf("Repo: %s\n", repo.FullName)
	}

	orgs, _, err := client.ListMyOrgs(ListOrgsOptions{ListOptions{Page: 1, PageSize: 10}})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing orgs: %v\n", err)
	}

	for _, org := range orgs {
		fmt.Printf("Org: %s\n", org.UserName)
	}
}
