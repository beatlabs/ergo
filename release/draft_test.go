package release_test

import (
	"errors"
	"testing"

	"github.com/taxibeat/ergo"
	"github.com/taxibeat/ergo/cli"
	"github.com/taxibeat/ergo/mock"
	"github.com/taxibeat/ergo/release"
)

func TestNewDraftShouldNotReturnNilObject(t *testing.T) {
	var host ergo.Host
	c := cli.NewCLI()
	releaseBranches := []string{"test1", "test2"}
	releaseBodyBranches := map[string]string{"test1": "foo", "test2": "bar"}

	if release.NewDraft(c, host, "test", "", releaseBranches, releaseBodyBranches) == nil {
		t.Error("expected draft object to not be nil.")
	}
}

// Create is responsible to create a new draft release.
func TestCreateShouldCreateDraft(t *testing.T) {
	host := &mock.RepositoryClient{}
	host.MockDiffCommitsFn = func() ([]*ergo.StatusReport, error) {
		statusRep1 := &ergo.StatusReport{
			Branch:     "release-xx",
			BaseBranch: "master",
			Ahead:      []*ergo.Commit{{Message: "aaa"}, {Message: "bbb"}},
			Behind:     []*ergo.Commit{{Message: "zzz"}},
		}
		statusRep2 := &ergo.StatusReport{
			Branch:     "release-zz",
			BaseBranch: "master",
			Ahead:      []*ergo.Commit{{Message: "aaa"}},
			Behind:     []*ergo.Commit{{Message: "zzz"}},
		}

		return []*ergo.StatusReport{statusRep1, statusRep2}, nil
	}
	host.MockCreateDraftReleaseFn = func() error {
		return nil
	}

	c := mock.CLI{}
	releaseBranches := []string{"release-xx", "release-zz"}
	releaseBodyBranches := map[string]string{"test1": "foo", "test2": "bar"}

	r := release.NewDraft(c, host, "test", "", releaseBranches, releaseBodyBranches)
	if r.Create(ctx, "test", "test") != nil {
		t.Error("expected create response not to be nil.")
	}
}

// Create is responsible to create a new draft release.
func TestCreateDiffCommitsShouldReturnError(t *testing.T) {
	host := &mock.RepositoryClient{}
	host.MockDiffCommitsFn = func() ([]*ergo.StatusReport, error) {
		return nil, errors.New("some error")
	}
	host.MockCreateDraftReleaseFn = func() error {
		return nil
	}

	c := mock.CLI{}
	releaseBranches := []string{"release-xx", "release-zz"}
	releaseBodyBranches := map[string]string{"test1": "foo", "test2": "bar"}

	r := release.NewDraft(c, host, "test", "", releaseBranches, releaseBodyBranches)
	if r.Create(ctx, "test", "test") == nil {
		t.Error("expected create response to be nil.")
	}
}

// Create is responsible to create a new draft release.
func TestCreateCreateDraftReleaseShouldReturnError(t *testing.T) {
	host := &mock.RepositoryClient{}
	host.MockDiffCommitsFn = func() ([]*ergo.StatusReport, error) {
		return []*ergo.StatusReport{}, nil
	}
	host.MockCreateDraftReleaseFn = func() error {
		return errors.New("some error")
	}

	c := mock.CLI{}
	releaseBranches := []string{"release-xx", "release-zz"}
	releaseBodyBranches := map[string]string{"test1": "foo", "test2": "bar"}

	r := release.NewDraft(c, host, "test", "", releaseBranches, releaseBodyBranches)
	if r.Create(ctx, "test", "test") == nil {
		t.Error("expected create response to be nil.")
	}
}
