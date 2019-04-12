package release_test

import (
	"errors"
	"testing"

	"github.com/taxibeat/ergo/mock"

	"github.com/taxibeat/ergo"
	"github.com/taxibeat/ergo/cli"
	"github.com/taxibeat/ergo/release"
)

func TestNewDeployShouldNotReturnNilObject(t *testing.T) {
	var host ergo.Host
	c := cli.NewCLI()

	deploy := release.NewDeploy(
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
	c := mock.CLI{}

	host.MockLastReleaseFn = func() (*ergo.Release, error) {
		return &ergo.Release{TagName: "1.0.0"}, nil
	}

	err := release.NewDeploy(
		c,
		host,
		"baseBranch",
		"",
		"",
		[]string{}, map[string]string{},
	).Do(ctx, "10ms", "1ms", false)

	if err != nil {
		t.Error("expected to not return error")
	}
}

func TestDoShouldReturnErrorOnLastRelease(t *testing.T) {
	host := &mock.RepositoryClient{}

	host.MockLastReleaseFn = func() (*ergo.Release, error) {
		return nil, errors.New("")
	}

	err := release.NewDeploy(
		nil,
		host,
		"baseBranch",
		"",
		"",
		[]string{}, map[string]string{},
	).Do(ctx, "10ms", "1ms", false)

	if err == nil {
		t.Error("expected to return error")
	}
}

func TestDoShouldReturnErrorOnConfirmation(t *testing.T) {
	host := &mock.RepositoryClient{}
	host.MockLastReleaseFn = func() (*ergo.Release, error) {
		return &ergo.Release{TagName: "1.0.0"}, nil
	}

	c := mock.CLI{MockConfirmation: func() (bool, error) {
		return false, errors.New("")
	}}

	err := release.NewDeploy(
		c,
		host,
		"baseBranch",
		"",
		"",
		[]string{}, map[string]string{},
	).Do(ctx, "10ms", "1ms", false)

	if err == nil {
		t.Error("expected to return error")
	}
}

func TestDoShouldNotReturnErrorWhenNotConfirm(t *testing.T) {
	host := &mock.RepositoryClient{}
	host.MockLastReleaseFn = func() (*ergo.Release, error) {
		return &ergo.Release{TagName: "1.0.0"}, nil
	}

	c := mock.CLI{MockConfirmation: func() (bool, error) {
		return false, nil
	}}

	err := release.NewDeploy(
		c,
		host,
		"baseBranch",
		"",
		"",
		[]string{}, map[string]string{},
	).Do(ctx, "10ms", "1ms", false)

	if err != nil {
		t.Error("expected not to return error")
	}
}

func TestDoShouldReturnErrorWhenReleaseTimeIsPast(t *testing.T) {
	host := &mock.RepositoryClient{}
	host.MockLastReleaseFn = func() (*ergo.Release, error) {
		return &ergo.Release{TagName: "1.0.0"}, nil
	}

	c := mock.CLI{MockConfirmation: func() (bool, error) {
		return true, nil
	}}

	err := release.NewDeploy(
		c,
		host,
		"baseBranch",
		"",
		"",
		[]string{}, map[string]string{},
	).Do(ctx, "1ms", "-1ms", false)

	if err == nil {
		t.Error("expected to return error")
	}
}

func TestDoShouldReturnErrorWithBadOffsetTime(t *testing.T) {
	host := &mock.RepositoryClient{}
	host.MockLastReleaseFn = func() (*ergo.Release, error) {
		return &ergo.Release{TagName: "1.0.0"}, nil
	}

	c := mock.CLI{MockConfirmation: func() (bool, error) {
		return true, nil
	}}

	err := release.NewDeploy(
		c,
		host,
		"baseBranch",
		"",
		"",
		[]string{}, map[string]string{},
	).Do(ctx, "1ms", "bad", false)

	if err == nil {
		t.Error("expected to return error")
	}
}

func TestDoShouldReleaseBranches(t *testing.T) {
	host := &mock.RepositoryClient{}
	host.MockLastReleaseFn = func() (*ergo.Release, error) {
		return &ergo.Release{TagName: "1.0.0"}, nil
	}
	host.MockUpdateBranchFromTagFn = func() error {
		return nil
	}
	host.MockLastReleaseFn = func() (*ergo.Release, error) {
		return &ergo.Release{TagName: "1.0.0"}, nil
	}
	host.MockEditReleaseFn = func() (*ergo.Release, error) {
		return &ergo.Release{TagName: "1.0.0"}, nil
	}

	c := mock.CLI{MockConfirmation: func() (bool, error) {
		return true, nil
	}}

	err := release.NewDeploy(
		c,
		host,
		"baseBranch",
		"suffix",
		"replace",
		[]string{"branch1", "branch2"},
		map[string]string{},
	).Do(ctx, "1ms", "1ms", false)

	if err != nil {
		t.Error("expected to not return error")
	}
}
