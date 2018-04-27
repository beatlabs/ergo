package main

import (
	"context"
	"fmt"

	"github.com/google/go-github/github"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

func main() {
	viper.AddConfigPath(".")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		fmt.Printf("error reading config file: %v", err)
		return
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: viper.GetString("github.accessToken")},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	// list all repositories for the authenticated user
	repo, resp, err := client.Repositories.Get(ctx, "taxibeat", "rest")
	if err != nil {
		fmt.Printf("error list repos: %v\n", err)
		return
	}

	fmt.Println(resp)
	fmt.Println(repo.Name)

	// tagName := "2018.04.19"
	// name := "April 19 2018"
	// isDraft := true
	// releaseBody := "this is a test release body created by ergo!"
	// release := &github.RepositoryRelease{
	// 	Name:    &name,
	// 	TagName: &tagName,
	// 	Draft:   &isDraft,
	// 	Body:    &releaseBody,
	// }
	// rel, resp, err := client.Repositories.CreateRelease(ctx, "taxibeat", "rest", release)
	// if err != nil {
	// 	fmt.Printf("error list repos: %v\n", err)
	// 	return
	// }

	// fmt.Println(resp)
	// fmt.Println(rel)

	repos, _, err := client.Repositories.ListByOrg(ctx, "taxibeat", &github.RepositoryListByOrgOptions{
		Type: "All",
	})
	if err != nil {
		fmt.Printf("error list repos: %v\n", err)
		return
	}
	for _, repo := range repos {
		fmt.Println(*(repo.IssuesURL))
	}
}
