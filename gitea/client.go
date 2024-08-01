package main

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"net/http"
	"os"
)

type GiteaUser struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	FullName  string `json:"full_name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

type Organization struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
}

type Repository struct {
	ID       int    `json:"id"`
	FullName string `json:"full_name"`
}

type ListOrgsOptions struct {
	Page    int `url:"page,omitempty"`
	PerPage int `url:"per_page,omitempty"`
}

type ListReposOptions struct {
	Page    int `url:"page,omitempty"`
	PerPage int `url:"per_page,omitempty"`
}

type Response struct {
	*http.Response
}

type GiteaClient struct {
	token string
	URL   string
	http  *http.Client
}

func NewGiteaClient(url, token string) *GiteaClient {
	return &GiteaClient{
		URL:   url,
		token: token,
		http:  &http.Client{},
	}
}

func (c *GiteaClient) GetUser() (*GiteaUser, error) {
	client := resty.NewWithClient(c.http)
	client.SetBaseURL(c.URL)
	client.SetAuthToken(c.token)

	resp, err := client.R().
		SetResult(&GiteaUser{}).
		Get("/user")

	if err != nil {
		return nil, fmt.Errorf("error fetching user information: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("error response from server: %s", resp.Status())
	}

	user := resp.Result().(*GiteaUser)
	return user, nil
}

func (c *GiteaClient) ListMyOrgs(opt ListOrgsOptions) ([]*Organization, *Response, error) {
	client := resty.NewWithClient(c.http)
	client.SetBaseURL(c.URL)
	client.SetAuthToken(c.token)

	resp, err := client.R().
		SetQueryParam("page", fmt.Sprintf("%d", opt.Page)).
		SetQueryParam("per_page", fmt.Sprintf("%d", opt.PerPage)).
		SetResult(&[]*Organization{}).
		Get("/user/orgs")

	if err != nil {
		return nil, nil, fmt.Errorf("error fetching organizations: %w", err)
	}

	if resp.IsError() {
		return nil, &Response{resp.RawResponse}, fmt.Errorf("error response from server: %s", resp.Status())
	}

	orgs := resp.Result().(*[]*Organization)
	return *orgs, &Response{resp.RawResponse}, nil
}

func (c *GiteaClient) ListMyRepos(opt ListReposOptions) ([]*Repository, *Response, error) {
	client := resty.NewWithClient(c.http)
	client.SetBaseURL(c.URL)
	client.SetAuthToken(c.token)

	resp, err := client.R().
		SetQueryParam("page", fmt.Sprintf("%d", opt.Page)).
		SetQueryParam("per_page", fmt.Sprintf("%d", opt.PerPage)).
		SetResult(&[]*Repository{}).
		Get("/user/repos")

	if err != nil {
		return nil, nil, fmt.Errorf("error fetching repositories: %w", err)
	}

	if resp.IsError() {
		return nil, &Response{resp.RawResponse}, fmt.Errorf("error response from server: %s", resp.Status())
	}

	repos := resp.Result().(*[]*Repository)
	return *repos, &Response{resp.RawResponse}, nil
}

func main() {
	baseURL := "https://gitcvt.hugpoint.tech/api/v1" // Ensure this URL is correct for your Gitea instance
	apiToken := os.Getenv("GITEA_API_TOKEN")         // Store your API token in an environment variable

	if apiToken == "" {
		fmt.Println("GITEA_API_TOKEN environment variable is not set")
		return
	}

	client := NewGiteaClient(baseURL, apiToken)

	// Test GetUser
	user, err := client.GetUser()
	if err != nil {
		fmt.Printf("Failed to fetch user information: %v\n", err)
		return
	}

	fmt.Printf("User ID: %d\n", user.ID)
	fmt.Printf("Login: %s\n", user.Login)
	fmt.Printf("Full Name: %s\n", user.FullName)
	fmt.Printf("Email: %s\n", user.Email)
	fmt.Printf("Avatar URL: %s\n", user.AvatarURL)

	// Test ListMyOrgs
	orgs, _, err := client.ListMyOrgs(ListOrgsOptions{Page: 1, PerPage: 10})
	if err != nil {
		fmt.Printf("Failed to fetch organizations: %v\n", err)
		return
	}

	fmt.Printf("Organizations:\n")
	for _, org := range orgs {
		fmt.Printf("- ID: %d, Login: %s\n", org.ID, org.Login)
	}

	// Test ListMyRepos
	repos, _, err := client.ListMyRepos(ListReposOptions{Page: 1, PerPage: 10})
	if err != nil {
		fmt.Printf("Failed to fetch repositories: %v\n", err)
		return
	}

	fmt.Printf("Repositories:\n")
	for _, repo := range repos {
		fmt.Printf("- ID: %d, FullName: %s\n", repo.ID, repo.FullName)
	}
}
