package github

import (
	"context"
	"fmt"

	"github.com/dbaltas/ergo/repo"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

//CreateDraftRelease creates a draft release
func CreateDraftRelease(ctx context.Context, accessToken string, organization string,
	repo string, name string, tagName string, releaseBody string) (*github.RepositoryRelease, error) {
	isDraft := true
	release := &github.RepositoryRelease{
		Name:    &name,
		TagName: &tagName,
		Draft:   &isDraft,
		Body:    &releaseBody,
	}
	client := githubClient(ctx, accessToken)
	release, _, err := client.Repositories.CreateRelease(ctx, organization, repo, release)

	return release, err
}

func LastRelease(ctx context.Context, accessToken string, organization string,
	repo string) (*github.RepositoryRelease, error) {
	client := githubClient(ctx, accessToken)
	release, _, err := client.Repositories.GetLatestRelease(ctx, organization, repo)

	return release, err
}

func EditRelease(ctx context.Context, accessToken string, organization string,
	repo string, release *github.RepositoryRelease) (*github.RepositoryRelease, error) {
	client := githubClient(ctx, accessToken)
	release, _, err := client.Repositories.EditRelease(ctx, organization, repo, *(release.ID), release)

	return release, err
}

func githubClient(ctx context.Context, accessToken string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}

//ReleaseBody output needed for github release body
func ReleaseBody(commitDiffBranches []repo.DiffCommitBranch, releaseBodyPrefix string) string {
	body := releaseBodyPrefix

	firstLinePrefix := "- [ ] "
	nextLinePrefix := "     "
	lineSeparator := "\r\n"

	for _, diffBranch := range commitDiffBranches {
		for _, commit := range diffBranch.Behind {
			body = fmt.Sprintf("%s%s%s",
				body,
				repo.FormatMessage(commit, firstLinePrefix, nextLinePrefix, lineSeparator),
				lineSeparator)
		}
	}
	return body
}
