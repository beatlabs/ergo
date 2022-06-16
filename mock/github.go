package mock

import (
	"context"

	"github.com/beatlabs/ergo"
)

// RepositoryClient is a mock implementation.
type RepositoryClient struct {
	MockCreateDraftReleaseFn  func() error
	MockLastReleaseFn         func() (*ergo.Release, error)
	MockEditReleaseFn         func() (*ergo.Release, error)
	MockPublishReleaseFn      func(ctx context.Context, releaseID int64) error
	MockCompareBranchFn       func() (*ergo.StatusReport, error)
	MockDiffCommitsFn         func() ([]*ergo.StatusReport, error)
	MockCreateTagFn           func() (*ergo.Tag, error)
	MockUpdateBranchFromTagFn func() error
	MockGetRefFn              func() (*ergo.Reference, error)
	MockGetRefFromTagFn       func() (*ergo.Reference, error)
	MockGetRepoNameFn         func() string
}

// CreateDraftRelease is a mock implementation.
func (r *RepositoryClient) CreateDraftRelease(ctx context.Context, name, tagName, releaseBody, targetBranch string) error {
	if r.MockCreateDraftReleaseFn != nil {
		return r.MockCreateDraftReleaseFn()
	}
	return nil
}

// LastRelease is a mock implementation.
func (r *RepositoryClient) LastRelease(ctx context.Context) (*ergo.Release, error) {
	if r.MockLastReleaseFn != nil {
		return r.MockLastReleaseFn()
	}
	return nil, nil
}

// EditRelease is a mock implementation.
func (r *RepositoryClient) EditRelease(ctx context.Context, release *ergo.Release) (*ergo.Release, error) {
	if r.MockEditReleaseFn != nil {
		return r.MockEditReleaseFn()
	}
	return nil, nil
}

// PublishRelease invokes the  mock implementation.
func (r *RepositoryClient) PublishRelease(ctx context.Context, releaseID int64) error {
	if r.MockPublishReleaseFn != nil {
		return r.MockPublishReleaseFn(ctx, releaseID)
	}
	return nil
}

// CompareBranch is a mock implementation.
func (r *RepositoryClient) CompareBranch(ctx context.Context, baseBranch, branch string) (*ergo.StatusReport, error) {
	if r.MockCompareBranchFn != nil {
		return r.MockCompareBranchFn()
	}
	return nil, nil
}

// DiffCommits is a mock implementation.
func (r *RepositoryClient) DiffCommits(ctx context.Context, releaseBranches []string, baseBranch string) ([]*ergo.StatusReport, error) {
	if r.MockDiffCommitsFn != nil {
		return r.MockDiffCommitsFn()
	}
	return nil, nil
}

// CreateTag is a mock implementation.
func (r *RepositoryClient) CreateTag(ctx context.Context, versionName, sha, m string) (*ergo.Tag, error) {
	if r.MockCreateTagFn != nil {
		return r.MockCreateTagFn()
	}
	return nil, nil
}

// UpdateBranchFromTag is a mock implementation.
func (r *RepositoryClient) UpdateBranchFromTag(ctx context.Context, tag, toBranch string, force bool) error {
	if r.MockUpdateBranchFromTagFn != nil {
		return r.MockUpdateBranchFromTagFn()
	}
	return nil
}

// GetRef is a mock implementation.
func (r *RepositoryClient) GetRef(ctx context.Context, branch string) (*ergo.Reference, error) {
	if r.MockGetRefFn != nil {
		return r.MockGetRefFn()
	}
	return nil, nil
}

// GetRefFromTag is a mock implementation.
func (r *RepositoryClient) GetRefFromTag(ctx context.Context, tag string) (*ergo.Reference, error) {
	if r.MockGetRefFromTagFn != nil {
		return r.MockGetRefFromTagFn()
	}
	return nil, nil
}

// GetRepoName is a mock implementation.
func (r *RepositoryClient) GetRepoName() string {
	if r.MockGetRepoNameFn != nil {
		return r.MockGetRepoNameFn()
	}
	return ""
}
