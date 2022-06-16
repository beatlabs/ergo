package release

import (
	"errors"
	"testing"

	"github.com/beatlabs/ergo"
	"github.com/beatlabs/ergo/cli"
	"github.com/beatlabs/ergo/mock"
)

func TestNewDraftShouldNotReturnNilObject(t *testing.T) {
	var host ergo.Host
	c := cli.NewCLI()
	releaseBranches := []string{"test1", "test2"}
	releaseBodyBranches := map[string]string{"test1": "foo", "test2": "bar"}

	if NewDraft(c, host, "test", "", releaseBranches, releaseBodyBranches) == nil {
		t.Error("expected draft object to not be nil.")
	}
}

// Create is responsible to create a new draft release.
func TestCreateShouldCreateDraft(t *testing.T) {
	host := &mock.RepositoryClient{}
	host.DiffCommitsFn = func() ([]*ergo.StatusReport, error) {
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
	host.CreateDraftReleaseFn = func() error {
		return nil
	}

	releaseBranches := []string{"release-xx", "release-zz"}
	releaseBodyBranches := map[string]string{"test1": "foo", "test2": "bar"}

	tests := map[string]struct {
		skipConfirmation      bool
		wantConfirmationCalls int
	}{
		"create draft asking for user confirmation": {skipConfirmation: false, wantConfirmationCalls: 1},
		"create draft skipping user confirmation":   {skipConfirmation: true, wantConfirmationCalls: 0},
	}
	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			c := &mock.CLI{}
			r := NewDraft(c, host, "test", "", releaseBranches, releaseBodyBranches)

			err := r.Create(ctx, "test", "test", tt.skipConfirmation)
			if err != nil {
				t.Errorf("Create returned error: %v", err)
			}
			if got, want := c.ConfirmationCalls, tt.wantConfirmationCalls; got != want {
				t.Errorf("NewDraft().Create(skipConfirmation=%t) -> confirmation calls=%d, want: %d", tt.skipConfirmation, got, want)
			}
		})
	}
}

// Create is responsible to create a new draft release.
func TestCreateDiffCommitsShouldReturnError(t *testing.T) {
	host := &mock.RepositoryClient{}
	host.DiffCommitsFn = func() ([]*ergo.StatusReport, error) {
		return nil, errors.New("some error")
	}
	host.CreateDraftReleaseFn = func() error {
		return nil
	}

	c := &mock.CLI{}
	releaseBranches := []string{"release-xx", "release-zz"}
	releaseBodyBranches := map[string]string{"test1": "foo", "test2": "bar"}

	r := NewDraft(c, host, "test", "", releaseBranches, releaseBodyBranches)
	if r.Create(ctx, "test", "test", false) == nil {
		t.Error("expected create response to be nil.")
	}
}

// Create is responsible to create a new draft release.
func TestCreateCreateDraftReleaseShouldReturnError(t *testing.T) {
	host := &mock.RepositoryClient{}
	host.DiffCommitsFn = func() ([]*ergo.StatusReport, error) {
		return []*ergo.StatusReport{}, nil
	}
	host.CreateDraftReleaseFn = func() error {
		return errors.New("some error")
	}

	c := &mock.CLI{}
	releaseBranches := []string{"release-xx", "release-zz"}
	releaseBodyBranches := map[string]string{"test1": "foo", "test2": "bar"}

	r := NewDraft(c, host, "test", "", releaseBranches, releaseBodyBranches)
	if r.Create(ctx, "test", "test", false) == nil {
		t.Error("expected create response to be nil.")
	}
}
