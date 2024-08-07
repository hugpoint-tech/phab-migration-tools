package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Helper function to create a mock server
func newMockServer(responseCode int, responseBody interface{}) *httptest.Server {
	handler := http.NewServeMux()
	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(responseCode)
		if responseBody != nil {
			respBody, _ := json.Marshal(responseBody)
			w.Write(respBody)
		}
	})
	return httptest.NewServer(handler)
}

// TestGetUser tests the GetUser method
func TestGetUser(t *testing.T) {
	mockUser := &User{
		ID:        1,
		UserName:  "testuser",
		FullName:  "Test User",
		Email:     "test@chsv.com",
		AvatarURL: "http://chsv.com/avatar",
	}

	mockServer := newMockServer(http.StatusOK, mockUser)
	defer mockServer.Close()

	client := &Client{
		url:         mockServer.URL,
		accessToken: "dummy-token",
		client:      &http.Client{},
	}

	user, err := client.GetUser()
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if user.UserName != "testuser" {
		t.Errorf("Expected username to be 'testuser', got '%s'", user.UserName)
	}
	if user.Email != "test@chsv.com" {
		t.Errorf("Expected email to be 'test@chsv.com', got '%s'", user.Email)
	}
}

// TestListMyRepos tests the ListMyRepos method
func TestListMyRepos(t *testing.T) {
	mockRepos := []*Repository{
		{
			ID:        1,
			FullName:  "repo1",
			CloneURL:  "http://chsv.com/repo1.git",
			Owner:     &User{UserName: "owner1"},
			Private:   false,
			HasIssues: true,
			HasWiki:   true,
			Created:   time.Now(),
			Updated:   time.Now(),
		},
	}

	mockServer := newMockServer(http.StatusOK, mockRepos)
	defer mockServer.Close()

	client := &Client{
		url:         mockServer.URL,
		accessToken: "dummy-token",
		client:      &http.Client{},
	}

	repos, _, err := client.ListMyRepos(ListReposOptions{ListOptions{Page: 1, PageSize: 10}})
	if err != nil {
		t.Fatalf("Failed to list repositories: %v", err)
	}

	if len(repos) != 1 {
		t.Fatalf("Expected 1 repository, got %d", len(repos))
	}
	if repos[0].FullName != "repo1" {
		t.Errorf("Expected repo full name to be 'repo1', got '%s'", repos[0].FullName)
	}
}

// TestListMyOrgs tests the ListMyOrgs method
func TestListMyOrgs(t *testing.T) {
	mockOrgs := []*Organization{
		{
			ID:          1,
			UserName:    "chsv1",
			FullName:    "Organization chsvs",
			Description: "Organization full of chsvs",
		},
	}

	mockServer := newMockServer(http.StatusOK, mockOrgs)
	defer mockServer.Close()

	client := &Client{
		url:         mockServer.URL,
		accessToken: "dummy-token",
		client:      &http.Client{},
	}

	orgs, _, err := client.ListMyOrgs(ListOrgsOptions{ListOptions{Page: 1, PageSize: 10}})
	if err != nil {
		t.Fatalf("Failed to list organizations: %v", err)
	}

	if len(orgs) != 1 {
		t.Fatalf("Expected 1 organization, got %d", len(orgs))
	}
	if orgs[0].UserName != "chsv1" {
		t.Errorf("Expected organization username to be 'org1', got '%s'", orgs[0].UserName)
	}
}
