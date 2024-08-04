package main

import (
	"os"
	"testing"
)

func TestGetUser(t *testing.T) {
	baseURL := "https://gitcvt.hugpoint.tech/api/v1" // Ensure this URL is correct for your Gitea instance
	apiToken := os.Getenv("GITEA_API_TOKEN")         // Store your API token in an environment variable

	if apiToken == "" {
		t.Fatal("GITEA_API_TOKEN environment variable is not set")
	}

	client := NewGiteaClient(baseURL, apiToken)
	user, err := client.GetUser()
	if err != nil {
		t.Fatalf("Failed to fetch user information: %v", err)
	}

	if user == nil {
		t.Fatal("User information is nil")
	}

	// Check some basic fields to ensure the user information is valid
	if user.Login == "" {
		t.Fatal("User login is empty")
	}

	if user.Email == "" {
		t.Fatal("User email is empty")
	}

	// Print user information for debugging purposes
	t.Logf("User ID: %d\n", user.ID)
	t.Logf("Login: %s\n", user.Login)
	t.Logf("Full Name: %s\n", user.FullName)
	t.Logf("Email: %s\n", user.Email)
	t.Logf("Avatar URL: %s\n", user.AvatarURL)
}

func TestListMyOrgs(t *testing.T) {
	baseURL := "https://gitcvt.hugpoint.tech/api/v1" // Ensure this URL is correct for your Gitea instance
	apiToken := os.Getenv("GITEA_API_TOKEN")         // Store your API token in an environment variable

	if apiToken == "" {
		t.Fatal("GITEA_API_TOKEN environment variable is not set")
	}

	client := NewGiteaClient(baseURL, apiToken)
	orgs, _, err := client.ListMyOrgs(ListOrgsOptions{Page: 1, PerPage: 10})
	if err != nil {
		t.Fatalf("Failed to fetch organizations: %v", err)
	}

	if orgs == nil {
		t.Fatal("Organizations list is nil")
	}

	if len(orgs) == 0 {
		t.Fatal("Organizations list is empty")
	}

	// Print organizations for debugging purposes
	for _, org := range orgs {
		t.Logf("Organization ID: %d, Login: %s\n", org.ID, org.Login)
	}
}

func TestListMyRepos(t *testing.T) {
	baseURL := "https://gitcvt.hugpoint.tech/api/v1" // Ensure this URL is correct for your Gitea instance
	apiToken := os.Getenv("GITEA_API_TOKEN")         // Store your API token in an environment variable

	if apiToken == "" {
		t.Fatal("GITEA_API_TOKEN environment variable is not set")
	}

	client := NewGiteaClient(baseURL, apiToken)
	repos, _, err := client.ListMyRepos(ListReposOptions{Page: 1, PerPage: 10})
	if err != nil {
		t.Fatalf("Failed to fetch repositories: %v", err)
	}

	if repos == nil {
		t.Fatal("Repositories list is nil")
	}

	if len(repos) == 0 {
		t.Fatal("Repositories list is empty")
	}

	// Print repositories for debugging purposes
	for _, repo := range repos {
		t.Logf("Repository ID: %d, FullName: %s\n", repo.ID, repo.FullName)
	}
}
