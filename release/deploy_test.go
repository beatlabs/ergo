package release

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/beatlabs/ergo/mock"

	"github.com/beatlabs/ergo"
	"github.com/beatlabs/ergo/cli"
)

func TestNewDeployShouldNotReturnNilObject(t *testing.T) {
	var host ergo.Host
	c := cli.NewCLI()

	deploy := NewDeploy(
		c,
		host,
		"baseBranch",
		"",
		"",
		[]string{}, map[string]string{},
	)
	if deploy == nil {
		t.Error("expected Deploy object to not be nil.")
	}
}

func TestDoShouldNotReturnErrorWithCorrectParameters(t *testing.T) {
	host := &mock.RepositoryClient{}
	c := &mock.CLI{}

	host.LastReleaseFn = func() (*ergo.Release, error) {
		return &ergo.Release{TagName: "1.0.0"}, nil
	}

	err := NewDeploy(
		c,
		host,
		"baseBranch",
		"",
		"",
		[]string{}, map[string]string{},
	).Do(ctx, "10ms", "1ms", false, false, false)

	if err != nil {
		t.Error("expected to not return error")
	}
}

func TestDoShouldReturnErrorOnLastRelease(t *testing.T) {
	host := &mock.RepositoryClient{}

	host.LastReleaseFn = func() (*ergo.Release, error) {
		return nil, errors.New("")
	}

	err := NewDeploy(
		nil,
		host,
		"baseBranch",
		"",
		"",
		[]string{}, map[string]string{},
	).Do(ctx, "10ms", "1ms", false, false, false)

	if err == nil {
		t.Error("expected to return error")
	}
}

func TestDoShouldReturnErrorOnConfirmation(t *testing.T) {
	host := &mock.RepositoryClient{}
	host.LastReleaseFn = func() (*ergo.Release, error) {
		return &ergo.Release{TagName: "1.0.0"}, nil
	}

	c := &mock.CLI{ConfirmationFn: func() (bool, error) {
		return false, errors.New("")
	}}

	err := NewDeploy(
		c,
		host,
		"baseBranch",
		"",
		"",
		[]string{}, map[string]string{},
	).Do(ctx, "10ms", "1ms", false, false, false)

	if err == nil {
		t.Error("expected to return error")
	}
}

func TestDoShouldNotReturnErrorWhenNotConfirm(t *testing.T) {
	host := &mock.RepositoryClient{}
	host.LastReleaseFn = func() (*ergo.Release, error) {
		return &ergo.Release{TagName: "1.0.0"}, nil
	}

	c := &mock.CLI{ConfirmationFn: func() (bool, error) {
		return false, nil
	}}

	err := NewDeploy(
		c,
		host,
		"baseBranch",
		"",
		"",
		[]string{}, map[string]string{},
	).Do(ctx, "10ms", "1ms", false, false, false)

	if err != nil {
		t.Error("expected not to return error")
	}
}

func TestDoShouldReturnErrorWhenReleaseTimeIsPast(t *testing.T) {
	host := &mock.RepositoryClient{}
	host.LastReleaseFn = func() (*ergo.Release, error) {
		return &ergo.Release{TagName: "1.0.0"}, nil
	}

	c := &mock.CLI{ConfirmationFn: func() (bool, error) {
		return true, nil
	}}

	err := NewDeploy(
		c,
		host,
		"baseBranch",
		"",
		"",
		[]string{}, map[string]string{},
	).Do(ctx, "1ms", "-1ms", false, false, false)

	if err == nil {
		t.Error("expected to return error")
	}
}

func TestDoShouldReturnErrorWithBadOffsetTime(t *testing.T) {
	host := &mock.RepositoryClient{}
	host.LastReleaseFn = func() (*ergo.Release, error) {
		return &ergo.Release{TagName: "1.0.0"}, nil
	}

	c := &mock.CLI{ConfirmationFn: func() (bool, error) {
		return true, nil
	}}

	err := NewDeploy(
		c,
		host,
		"baseBranch",
		"",
		"",
		[]string{}, map[string]string{},
	).Do(ctx, "1ms", "bad", false, false, false)

	if err == nil {
		t.Error("expected to return error")
	}
}

func TestDoShouldReleaseBranches(t *testing.T) {
	host := &mock.RepositoryClient{}
	host.LastReleaseFn = func() (*ergo.Release, error) {
		return &ergo.Release{TagName: "1.0.0"}, nil
	}
	host.UpdateBranchFromTagFn = func() error {
		return nil
	}
	host.LastReleaseFn = func() (*ergo.Release, error) {
		return &ergo.Release{TagName: "1.0.0"}, nil
	}
	host.EditReleaseFn = func() (*ergo.Release, error) {
		return &ergo.Release{TagName: "1.0.0"}, nil
	}

	c := &mock.CLI{ConfirmationFn: func() (bool, error) {
		return true, nil
	}}

	err := NewDeploy(
		c,
		host,
		"baseBranch",
		"suffix",
		"replace",
		[]string{"branch1", "branch2"},
		map[string]string{},
	).Do(ctx, "1ms", "1ms", false, false, false)

	if err != nil {
		t.Error("expected to not return error")
	}
}

func TestDoShouldDeployWhenSkippingUserConfirmation(t *testing.T) {
	tests := map[string]struct {
		skipConfirmation      bool
		wantConfirmationCalls int
	}{
		"NewDeploy().Do asking for user confirmation": {skipConfirmation: false, wantConfirmationCalls: 1},
		"NewDeploy().Do skipping user confirmation":   {skipConfirmation: true, wantConfirmationCalls: 0},
	}
	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			c := &mock.CLI{}
			host := &mock.RepositoryClient{
				LastReleaseFn: func() (*ergo.Release, error) {
					return &ergo.Release{TagName: "1.0.0"}, nil
				},
			}
			err := NewDeploy(
				c,
				host,
				"baseBranch",
				"suffix",
				"replace",
				[]string{"branch1", "branch2"},
				map[string]string{},
			).Do(ctx, "1ms", "1ms", false, tt.skipConfirmation, false)
			if err != nil {
				t.Errorf("NewDeploy().Do(skipConfirmation=%t) returned error: %v", tt.skipConfirmation, err)
			}
			if got, want := c.ConfirmationCalls, tt.wantConfirmationCalls; got != want {
				t.Errorf("NewDeploy().Do(skipConfirmation=%t) -> confirmation calls=%d, want: %d", tt.skipConfirmation, got, want)
			}
		})
	}
}

func TestDoWithPublishDraftEnabledSuccess(t *testing.T) {
	c := &mock.CLI{}
	host := &mock.RepositoryClient{
		LastReleaseFn: func() (*ergo.Release, error) {
			return &ergo.Release{TagName: "1.0.0", Draft: true}, nil
		},
	}
	err := NewDeploy(
		c,
		host,
		"baseBranch",
		"suffix",
		"replace",
		[]string{"branch1", "branch2"},
		map[string]string{},
	).Do(ctx, "1ms", "1ms", false, false, true)
	if err != nil {
		t.Errorf("NewDeploy().Do(publishDraft=true) returned error: %v", err)
	}
}

func TestDoWithPublishDraftEnabledError(t *testing.T) {
	tests := map[string]struct {
		LastReleaseFn    func() (*ergo.Release, error)
		PublishReleaseFn func(ctx context.Context, releaseID int64) error
	}{
		"last release is not a draft": {
			LastReleaseFn: func() (*ergo.Release, error) {
				return &ergo.Release{TagName: "1.0.0", Draft: false}, nil
			},
		},
		"publish release returns error": {
			LastReleaseFn: func() (*ergo.Release, error) {
				return &ergo.Release{TagName: "1.0.0", Draft: true}, nil
			},
			PublishReleaseFn: func(ctx context.Context, releaseID int64) error {
				return fmt.Errorf("something went wrong")
			},
		},
	}
	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			c := &mock.CLI{}
			host := &mock.RepositoryClient{
				LastReleaseFn:    tt.LastReleaseFn,
				PublishReleaseFn: tt.PublishReleaseFn,
			}
			err := NewDeploy(
				c,
				host,
				"baseBranch",
				"suffix",
				"replace",
				[]string{"branch1", "branch2"},
				map[string]string{},
			).Do(ctx, "1ms", "1ms", false, false, true)
			if err == nil {
				t.Errorf("NewDeploy().Do(publishDraft=true) when %q should return error", testName)
			}
		})
	}
}
