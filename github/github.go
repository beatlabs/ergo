package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/thebeatapp/ergo"
	"golang.org/x/oauth2"
)

// RepositoryClient for Github API.
type RepositoryClient struct {
	organization string
	repo         string
	client       *github.Client
	curRelease   *github.RepositoryRelease
}

// NewGithubClient set up a github client.
func NewGithubClient(ctx context.Context, accessToken string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}

// NewRepositoryClient instantiate a RepositoryClient.
func NewRepositoryClient(organization, repo string, client *github.Client) *RepositoryClient {
	return &RepositoryClient{
		organization: organization,
		repo:         repo,
		client:       client,
	}
}

// CreateDraftRelease creates a draft release.
func (gc *RepositoryClient) CreateDraftRelease(ctx context.Context, name, tagName, releaseBody string) error {
	isDraft := true
	githubRelease := &github.RepositoryRelease{
		Name:    &name,
		TagName: &tagName,
		Draft:   &isDraft,
		Body:    &releaseBody,
	}

	githubRelease, _, err := gc.client.Repositories.CreateRelease(
		ctx,
		gc.organization,
		gc.repo,
		githubRelease,
	)

	gc.curRelease = githubRelease

	return err
}

// LastRelease fetches the latest release for a repository.
func (gc *RepositoryClient) LastRelease(ctx context.Context) (*ergo.Release, error) {
	githubRelease, _, err := gc.client.Repositories.GetLatestRelease(
		ctx, gc.organization, gc.repo)

	errResponse, ok := err.(*github.ErrorResponse)

	errHasResponse := ok
	if errHasResponse && errResponse.Response.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	gc.curRelease = githubRelease

	return &ergo.Release{
		ID:         *githubRelease.ID,
		Body:       *githubRelease.Body,
		TagName:    githubRelease.GetTagName(),
		ReleaseURL: githubRelease.GetHTMLURL(),
	}, nil
}

// EditRelease allows to edit a repository release.
func (gc *RepositoryClient) EditRelease(ctx context.Context, release *ergo.Release) (*ergo.Release, error) {

	if gc.curRelease == nil {
		return nil, errors.New("curRelease is empty")
	}

	githubRelease := gc.curRelease
	githubRelease.Body = &release.Body

	githubRelease, _, err := gc.client.Repositories.EditRelease(
		ctx, gc.organization, gc.repo, release.ID, githubRelease)

	release.Body = *githubRelease.Body

	return release, err
}

// GetRef branch reference object given the branch name.
func (gc *RepositoryClient) GetRef(ctx context.Context, branch string) (*ergo.Reference, error) {
	ref, _, err := gc.client.Git.GetRef(ctx, gc.organization, gc.repo, "refs/heads/"+branch)
	if err != nil {
		return nil, err
	}

	return &ergo.Reference{SHA: *ref.Object.SHA, Ref: *ref.Ref}, nil
}

// CreateTag given the version name.
func (gc *RepositoryClient) CreateTag(ctx context.Context, versionName, sha, m string) (*ergo.Tag, error) {
	s := "commit"
	tag := github.Tag{
		Tag:     &versionName,
		Message: &m,
		Object:  &github.GitObject{Type: &s, SHA: &sha},
	}
	t, _, err := gc.client.Git.CreateTag(ctx, gc.organization, gc.repo, &tag)
	errResponse, ok := err.(*github.ErrorResponse)
	if ok && errResponse.Response.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error on tag creation: %v", err)
	}

	url := "tags/" + versionName
	ref := github.Reference{
		Object: &github.GitObject{
			SHA: &sha,
		},
		Ref: &url,
	}

	_, _, err = gc.client.Git.CreateRef(ctx, gc.organization, gc.repo, &ref)
	errResponse, ok = err.(*github.ErrorResponse)
	if ok && errResponse.Response.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error on tag creation: %v", err)
	}

	return &ergo.Tag{Name: *t.Tag}, err
}

// CompareBranch compare the base branch with the given one.
func (gc *RepositoryClient) CompareBranch(ctx context.Context, baseBranch, branch string) (*ergo.StatusReport, error) {
	commitsAhead, err := gc.commitsDiff(ctx, baseBranch, branch)
	if err != nil {
		return nil, err
	}

	commitsBehind, err := gc.commitsDiff(ctx, branch, baseBranch)
	if err != nil {
		return nil, err
	}

	return &ergo.StatusReport{Branch: branch, BaseBranch: baseBranch, Ahead: commitsAhead, Behind: commitsBehind}, nil
}

// commitsDiff finds the differences in commits between two branches.
func (gc *RepositoryClient) commitsDiff(ctx context.Context, baseBranch, branch string) ([]*ergo.Commit, error) {
	comparison, _, err := gc.client.Repositories.CompareCommits(ctx, gc.organization, gc.repo, baseBranch, branch)
	if err != nil {
		return nil, err
	}

	var commitsAhead []*ergo.Commit
	for _, commit := range comparison.Commits {
		commitAhead := &ergo.Commit{Message: *commit.Commit.Message}
		commitsAhead = append(commitsAhead, commitAhead)
	}

	return commitsAhead, nil
}

// DiffCommits is responsible to find the diff-commits and return a StatusReport for each of
// given releaseBranches.
func (gc *RepositoryClient) DiffCommits(ctx context.Context, releaseBranches []string, baseBranch string) ([]*ergo.StatusReport, error) {
	var statusReports []*ergo.StatusReport
	for _, branch := range releaseBranches {
		statusReport, err := gc.CompareBranch(ctx, baseBranch, branch)
		if err != nil {
			return nil, fmt.Errorf("error comparing base branch %s %s:%s", baseBranch, branch, err)
		}
		statusReports = append(statusReports, statusReport)
	}
	return statusReports, nil
}

// UpdateBranchFromTag is responsible to update a branch from tag.
func (gc *RepositoryClient) UpdateBranchFromTag(ctx context.Context, tag, toBranch string, force bool) error {
	ref, err := gc.getRefFromGitHub(ctx, tag)
	if err != nil {
		return err
	}

	branchRef := "heads/" + toBranch
	ref.Ref = &branchRef
	_, _, err = gc.client.Git.UpdateRef(ctx, gc.organization, gc.repo, ref, force)
	if err != nil {
		return fmt.Errorf("error on update branch from tag: %v", err)
	}

	return nil
}

// GetRefFromTag get reference from tag.
func (gc *RepositoryClient) GetRefFromTag(ctx context.Context, tag string) (*ergo.Reference, error) {
	ref, err := gc.getRefFromGitHub(ctx, tag)
	if err != nil {
		return nil, err
	}
	if ref == nil {
		return nil, nil
	}
	return &ergo.Reference{SHA: *ref.Object.SHA, Ref: *ref.Ref}, nil
}

// getRefFromGitHub get reference from github and returns the github.Reference object.
func (gc *RepositoryClient) getRefFromGitHub(ctx context.Context, tag string) (*github.Reference, error) {
	ref, _, err := gc.client.Git.GetRef(ctx, gc.organization, gc.repo, "tags/"+tag)

	errResponse, ok := err.(*github.ErrorResponse)
	if ok && errResponse.Response.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting tag reference %v", err)
	}

	return ref, nil
}

// GetRepoName return the repository name.
func (gc *RepositoryClient) GetRepoName() string {
	return gc.organization + "/" + gc.repo
}
