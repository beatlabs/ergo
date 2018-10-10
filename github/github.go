package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/dbaltas/ergo/repo"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// Client for github API
type Client struct {
	ctx          context.Context
	accessToken  string
	organization string
	repo         string
	client       *github.Client
}

// NewClient instantiate a Client
func NewClient(ctx context.Context, accessToken, organization, repo string) (*Client, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("github.access_token not defined in config")
	}

	client := githubClient(ctx, accessToken)
	return &Client{
		ctx:          ctx,
		accessToken:  accessToken,
		organization: organization,
		repo:         repo,
		client:       client,
	}, nil
}

// CreateDraftRelease creates a draft release.
func (gc *Client) CreateDraftRelease(name, tagName, releaseBody string) (*github.RepositoryRelease, error) {
	isDraft := true
	release := &github.RepositoryRelease{
		Name:    &name,
		TagName: &tagName,
		Draft:   &isDraft,
		Body:    &releaseBody,
	}

	release, _, err := gc.client.Repositories.CreateRelease(
		gc.ctx,
		gc.organization,
		gc.repo,
		release,
	)

	return release, err
}

// LastRelease fetches the latest release for a repository.
func (gc *Client) LastRelease() (*github.RepositoryRelease, error) {
	release, _, err := gc.client.Repositories.GetLatestRelease(
		gc.ctx, gc.organization, gc.repo)

	return release, err
}

// EditRelease allows to edit a repository release.
func (gc *Client) EditRelease(release *github.RepositoryRelease) (*github.RepositoryRelease, error) {
	release, _, err := gc.client.Repositories.EditRelease(
		gc.ctx, gc.organization, gc.repo, *(release.ID), release)

	return release, err
}

// CreatePR creates a pull request
func (gc *Client) CreatePR(baseBranch, compareBranch, title, body string) (*github.PullRequest, error) {
	pull := &github.NewPullRequest{
		Title: &title,
		Head:  &compareBranch,
		Base:  &baseBranch,
		Body:  &body,
	}

	pr, _, err := gc.client.PullRequests.Create(gc.ctx, gc.organization, gc.repo, pull)
	if err != nil {
		return nil, err
	}

	return pr, nil
}

// GetPR gets a pull request
func (gc *Client) GetPR(number int) (*github.PullRequest, error) {
	pr, _, err := gc.client.PullRequests.Get(gc.ctx, gc.organization, gc.repo, number)
	if err != nil {
		return nil, err
	}

	return pr, nil
}

// RequestReviewersForPR assigns reviewers to a pull request
func (gc *Client) RequestReviewersForPR(number int, reviewers, teamReviewers string) (*github.PullRequest, error) {
	payload := github.ReviewersRequest{
		Reviewers:     strings.Split(reviewers, ","),
		TeamReviewers: strings.Split(teamReviewers, ","),
	}
	fmt.Println(github.Stringify(payload))
	pr, _, err := gc.client.PullRequests.RequestReviewers(gc.ctx, gc.organization, gc.repo, number, payload)
	if err != nil {
		return nil, err
	}

	return pr, nil
}

// ListPRs creates a pull request
func (gc *Client) ListPRs() ([]*github.PullRequest, error) {
	opt := &github.PullRequestListOptions{
		Sort:      "created",
		Direction: "desc",
	}

	pulls, _, err := gc.client.PullRequests.List(gc.ctx, gc.organization, gc.repo, opt)
	if err != nil {
		return nil, err
	}

	return pulls, nil
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

func githubClient(ctx context.Context, accessToken string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}
