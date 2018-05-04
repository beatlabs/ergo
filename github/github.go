package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/dbaltas/ergo/repo"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// CreateDraftRelease creates a draft release.
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

// LastRelease fetches the latest release for a repository.
func LastRelease(ctx context.Context, accessToken string, organization string,
	repo string) (*github.RepositoryRelease, error) {
	client := githubClient(ctx, accessToken)
	release, _, err := client.Repositories.GetLatestRelease(ctx, organization, repo)

	return release, err
}

// EditRelease allows to edit a repository release.
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

// ReleaseBody output needed for github release body.
func ReleaseBody(commitDiffBranches []repo.DiffCommitBranch, releaseBodyPrefix string, branchMap map[string]string) string {
	var formattedCommits []string
	var formattedBranches []string
	var header, body string

	firstLinePrefix := "- [ ] "
	nextLinePrefix := "     "
	lineSeparator := "\r\n"

	for _, diffBranch := range commitDiffBranches {
		branchText, ok := branchMap[diffBranch.Branch]
		if !ok {
			branchText = branchMap[diffBranch.Branch]
		}
		formattedBranches = append(formattedBranches,
			fmt.Sprintf("%s ![](https://img.shields.io/badge/released-No-red.svg)", branchText))
	}

	for _, commit := range commitDiffBranches[0].Behind {
		formattedCommits = append(formattedCommits, repo.FormatMessage(commit, firstLinePrefix, nextLinePrefix, lineSeparator))
		body = fmt.Sprintf("%s%s%s",
			body,
			repo.FormatMessage(commit, firstLinePrefix, nextLinePrefix, lineSeparator),
			lineSeparator)
	}

	header = strings.Join(formattedBranches, " ")
	body = strings.Join(formattedCommits, lineSeparator)
	parts := []string{header, releaseBodyPrefix, body}

	return strings.Join(parts, strings.Repeat(lineSeparator, 2))
}
