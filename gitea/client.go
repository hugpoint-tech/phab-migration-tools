package main

import (
	"code.gitea.io/sdk/gitea"
	"fmt"
	"log"
	"os"
	"testing"
)

func main() {
	URL := "https://gitcvt.hugpoint.tech"
	token := os.Getenv("GITEA_TOKEN")

	// Create new Gitea client
	client, err := gitea.NewClient(URL, gitea.SetToken(token))
	if err != nil {
		log.Fatalf("Failed to create Gitea client: %v", err)
	}

	// Fetch user information
	user, _, err := client.GetMyUserInfo()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching user: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("User: %s\n", user.UserName)
	fmt.Printf("Full Name: %s\n", user.FullName)
	fmt.Printf("Email: %s\n", user.Email)

	// List repositories
	repos, _, err := client.ListMyRepos(gitea.ListReposOptions{
		ListOptions: gitea.ListOptions{
			Page:     1,
			PageSize: 10},
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing repos: %v\n", err)
	}

	if len(repos) == 0 {
		fmt.Println("No repositories found.")
	} else {
		for _, repo := range repos {
			fmt.Printf("Repo: %s\n", repo.FullName)
		}
	}

	// List organizations
	orgs, _, err := client.ListMyOrgs(gitea.ListOrgsOptions{
		ListOptions: gitea.ListOptions{
			Page:     1,
			PageSize: 10},
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing orgs: %v\n", err)
	}

	if len(orgs) == 0 {
		fmt.Println("No organizations found.")
	} else {
		for _, org := range orgs {
			fmt.Printf("Org: %s\n", org.UserName)
		}
	}

	// List teams that the authenticated user is a member of
	teams, _, err := client.ListMyTeams(&gitea.ListTeamsOptions{
		ListOptions: gitea.ListOptions{
			Page:     1,
			PageSize: 10,
		},
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching teams: %v\n", err)
	}

	if len(teams) == 0 {
		fmt.Println("No teams found.")
	} else {
		for _, team := range teams {
			fmt.Printf("Team ID: %d, Name: %s, Description: %s, Organization: %s\n",
				team.ID, team.Name, team.Description, team.Organization.UserName)
		}
	}

	// List followers
	followers, _, err := client.ListMyFollowers(gitea.ListFollowersOptions{})
	if err != nil {
		log.Fatalf("Failed to list followers: %v", err)
	}

	if len(orgs) == 0 {
		fmt.Println("No followers found.")
	} else {
		for _, follower := range followers {
			fmt.Printf("Follower: %s\n", follower.UserName)
		}
	}
}

func NewGiteaClient(t *testing.T) *gitea.Client {
	apiToken := os.Getenv("GITEA_TOKEN")
	if apiToken == "" {
		t.Fatal("GITEA_TOKEN environment variable is not set")
	}

	client, err := gitea.NewClient("https://gitcvt.hugpoint.tech", gitea.SetToken(apiToken))
	if err != nil {
		t.Fatalf("Error creating Gitea client: %v", err)
	}

	return client
}
