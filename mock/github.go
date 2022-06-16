package mock

import (
	"context"

	"github.com/beatlabs/ergo"
)

// RepositoryClient is a mock implementation.
type RepositoryClient struct {
	CreateDraftReleaseFn  func() error
	LastReleaseFn         func() (*ergo.Release, error)
	EditReleaseFn         func() (*ergo.Release, error)
	PublishReleaseFn      func(ctx context.Context, releaseID int64) error
	CompareBranchFn       func() (*ergo.StatusReport, error)
	DiffCommitsFn         func() ([]*ergo.StatusReport, error)
	CreateTagFn           func() (*ergo.Tag, error)
	UpdateBranchFromTagFn func() error
	GetRefFn              func() (*ergo.Reference, error)
	GetRefFromTagFn       func() (*ergo.Reference, error)
	GetRepoNameFn         func() string
}

// CreateDraftRelease is a mock implementation.
func (r *RepositoryClient) CreateDraftRelease(ctx context.Context, name, tagName, releaseBody, targetBranch string) error {
	if r.CreateDraftReleaseFn != nil {
		return r.CreateDraftReleaseFn()
	}
	return nil
}

// LastRelease is a mock implementation.
func (r *RepositoryClient) LastRelease(ctx context.Context) (*ergo.Release, error) {
	if r.LastReleaseFn != nil {
		return r.LastReleaseFn()
	}
	return nil, nil
}

// EditRelease is a mock implementation.
func (r *RepositoryClient) EditRelease(ctx context.Context, release *ergo.Release) (*ergo.Release, error) {
	if r.EditReleaseFn != nil {
		return r.EditReleaseFn()
	}
	return nil, nil
}

// PublishRelease invokes the  mock implementation.
func (r *RepositoryClient) PublishRelease(ctx context.Context, releaseID int64) error {
	if r.PublishReleaseFn != nil {
		return r.PublishReleaseFn(ctx, releaseID)
	}
	return nil
}

// CompareBranch is a mock implementation.
func (r *RepositoryClient) CompareBranch(ctx context.Context, baseBranch, branch string) (*ergo.StatusReport, error) {
	if r.CompareBranchFn != nil {
		return r.CompareBranchFn()
	}
	return nil, nil
}

// DiffCommits is a mock implementation.
func (r *RepositoryClient) DiffCommits(ctx context.Context, releaseBranches []string, baseBranch string) ([]*ergo.StatusReport, error) {
	if r.DiffCommitsFn != nil {
		return r.DiffCommitsFn()
	}
	return nil, nil
}

// CreateTag is a mock implementation.
func (r *RepositoryClient) CreateTag(ctx context.Context, versionName, sha, m string) (*ergo.Tag, error) {
	if r.CreateTagFn != nil {
		return r.CreateTagFn()
	}
	return nil, nil
}

// UpdateBranchFromTag is a mock implementation.
func (r *RepositoryClient) UpdateBranchFromTag(ctx context.Context, tag, toBranch string, force bool) error {
	if r.UpdateBranchFromTagFn != nil {
		return r.UpdateBranchFromTagFn()
	}
	return nil
}

// GetRef is a mock implementation.
func (r *RepositoryClient) GetRef(ctx context.Context, branch string) (*ergo.Reference, error) {
	if r.GetRefFn != nil {
		return r.GetRefFn()
	}
	return nil, nil
}

// GetRefFromTag is a mock implementation.
func (r *RepositoryClient) GetRefFromTag(ctx context.Context, tag string) (*ergo.Reference, error) {
	if r.GetRefFromTagFn != nil {
		return r.GetRefFromTagFn()
	}
	return nil, nil
}

// GetRepoName is a mock implementation.
func (r *RepositoryClient) GetRepoName() string {
	if r.GetRepoNameFn != nil {
		return r.GetRepoNameFn()
	}
	return ""
}
