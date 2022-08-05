package release

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/beatlabs/ergo"
	"github.com/beatlabs/ergo/cli"
	"github.com/beatlabs/ergo/mock"
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

func TestNonLinearIntervals(t *testing.T) {
	tests := []struct {
		name      string
		branches  []string
		intervals string
	}{
		{
			name:      "single branch, single interval",
			branches:  []string{"b1"},
			intervals: "1ms",
		},
		{
			name:      "multiple branches, single interval",
			branches:  []string{"b1", "b2", "b3", "b4"},
			intervals: "1ms",
		},
		{
			name:      "multiple branches, nonlinear intervals",
			branches:  []string{"b1", "b2", "b3", "b4"},
			intervals: "10ms,5ms,1ms,1ms",
		},
		{
			name:      "multiple branches, fewer intervals",
			branches:  []string{"b1", "b2", "b3", "b4"},
			intervals: "10ms,5ms",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
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
				test.branches, map[string]string{},
			).Do(ctx, test.intervals, "1ms", false, false, false)

			if err != nil {
				t.Errorf("NewDeploy().Do() returned error: %v", err)
			}
		})
	}
}

func TestPrintReleaseTimeBoard_NonLinearIntervalHandling(t *testing.T) {
	ts := func(t time.Time) string {
		return t.Format("15:04 MST")
	}
	tests := []struct {
		name              string
		branches          []string
		intervals         string
		expectedPrintRows [][]string
	}{
		{
			name:      "single branch, single interval",
			branches:  []string{"branch1"},
			intervals: "1m",
			expectedPrintRows: [][]string{
				{"branch1", ts(time.Now())},
			},
		},
		{
			name:      "multiple branches, single interval",
			branches:  []string{"branch1", "branch2", "branch3", "branch4"},
			intervals: "1m",
			expectedPrintRows: [][]string{
				{"branch1", ts(time.Now())},
				{"branch2", ts(time.Now().Add(1 * time.Minute))},
				{"branch3", ts(time.Now().Add(2 * time.Minute))},
				{"branch4", ts(time.Now().Add(3 * time.Minute))},
			},
		},
		{
			name:      "multiple branches, nonlinear intervals",
			branches:  []string{"branch1", "branch2", "branch3", "branch4"},
			intervals: "10m,5m,1m",
			expectedPrintRows: [][]string{
				{"branch1", ts(time.Now())},
				{"branch2", ts(time.Now().Add(10 * time.Minute))},
				{"branch3", ts(time.Now().Add(15 * time.Minute))},
				{"branch4", ts(time.Now().Add(16 * time.Minute))},
			},
		},
		{
			name:      "multiple branches, fewer intervals",
			branches:  []string{"branch1", "branch2", "branch3", "branch4"},
			intervals: "10m,5m",
			expectedPrintRows: [][]string{
				{"branch1", ts(time.Now())},
				{"branch2", ts(time.Now().Add(10 * time.Minute))},
				{"branch3", ts(time.Now().Add(15 * time.Minute))},
				{"branch4", ts(time.Now().Add(25 * time.Minute))},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cliMock := &mock.CLI{}
			deploy := &Deploy{
				c:               cliMock,
				releaseBranches: test.branches,
			}
			intervalDurations, releaseTimer, err := deploy.calculateReleaseTime(test.intervals, "1ms")
			if err != nil {
				t.Errorf("NewDeploy().Do() returned error: %v", err)
			}

			releaseTime := *releaseTimer

			deploy.printReleaseTimeBoard(releaseTime, deploy.releaseBranches, intervalDurations)
			if len(cliMock.PrintTableCalls) != 1 {
				t.Errorf("expected exactly one interaction with PrintTable")
			}
			PrintTableCall := cliMock.PrintTableCalls[0]
			expected := []string{"Branch", "Start Time"}
			if !reflect.DeepEqual(expected, PrintTableCall.Header) {
				t.Errorf("expected %v to equal %v", expected, PrintTableCall.Header)
			}
			if !reflect.DeepEqual(test.expectedPrintRows, PrintTableCall.Values) {
				t.Errorf("expected %v to equal %v", test.expectedPrintRows, PrintTableCall.Values)
			}
		})
	}
}

func TestDeployToAllReleaseBranches_NonLinearIntervalHandling(t *testing.T) {
	ts := func(t time.Time) string {
		return t.Format("15:04:05")
	}
	releaseTime := time.Date(2022, 8, 4, 13, 37, 0, 0, time.UTC)
	tests := []struct {
		name              string
		branches          []string
		intervals         string
		expectedPrints    []string
		expectedFinalTime time.Time
	}{
		{
			name:      "single branch, single interval",
			branches:  []string{"branch1"},
			intervals: "10m",
			expectedPrints: []string{
				fmt.Sprintf("Deploying %v branch1", ts(releaseTime)),
				fmt.Sprintf("%v Triggered Successfully", ts(releaseTime)),
			},
			expectedFinalTime: releaseTime,
		},
		{
			name:      "multiple branches, single interval",
			branches:  []string{"branch1", "branch2"},
			intervals: "10m",
			expectedPrints: []string{
				fmt.Sprintf("Deploying %v branch1", ts(releaseTime)),
				fmt.Sprintf("%v Triggered Successfully", ts(releaseTime)),
				fmt.Sprintf("Deploying %v branch2", ts(releaseTime.Add(10*time.Minute))),
				fmt.Sprintf("%v Triggered Successfully", ts(releaseTime.Add(10*time.Minute))),
			},
			expectedFinalTime: releaseTime.Add(10 * time.Minute),
		},
		{
			name:      "multiple branches, decreasing interval",
			branches:  []string{"branch1", "branch2", "branch3", "branch4", "branch5"},
			intervals: "10m,5m,1m,1m",
			expectedPrints: []string{
				fmt.Sprintf("Deploying %v branch1", ts(releaseTime)),
				fmt.Sprintf("%v Triggered Successfully", ts(releaseTime)),
				fmt.Sprintf("Deploying %v branch2", ts(releaseTime.Add(10*time.Minute))), // 10
				fmt.Sprintf("%v Triggered Successfully", ts(releaseTime.Add(10*time.Minute))),
				fmt.Sprintf("Deploying %v branch3", ts(releaseTime.Add(15*time.Minute))), // 10 + 5
				fmt.Sprintf("%v Triggered Successfully", ts(releaseTime.Add(15*time.Minute))),
				fmt.Sprintf("Deploying %v branch4", ts(releaseTime.Add(16*time.Minute))), // 10 + 5 + 1
				fmt.Sprintf("%v Triggered Successfully", ts(releaseTime.Add(16*time.Minute))),
				fmt.Sprintf("Deploying %v branch5", ts(releaseTime.Add(17*time.Minute))), // 10 + 5 + 1 + 1
				fmt.Sprintf("%v Triggered Successfully", ts(releaseTime.Add(17*time.Minute))),
			},
			expectedFinalTime: releaseTime.Add(17 * time.Minute),
		},
		{
			name:      "multiple branches, fewer intervals",
			branches:  []string{"branch1", "branch2", "branch3", "branch4", "branch5"},
			intervals: "10m,5m",
			expectedPrints: []string{
				fmt.Sprintf("Deploying %v branch1", ts(releaseTime)),
				fmt.Sprintf("%v Triggered Successfully", ts(releaseTime)),
				fmt.Sprintf("Deploying %v branch2", ts(releaseTime.Add(10*time.Minute))), // 10
				fmt.Sprintf("%v Triggered Successfully", ts(releaseTime.Add(10*time.Minute))),
				fmt.Sprintf("Deploying %v branch3", ts(releaseTime.Add(15*time.Minute))), // 10 + 5
				fmt.Sprintf("%v Triggered Successfully", ts(releaseTime.Add(15*time.Minute))),
				fmt.Sprintf("Deploying %v branch4", ts(releaseTime.Add(25*time.Minute))), // 10 + 5 + 10
				fmt.Sprintf("%v Triggered Successfully", ts(releaseTime.Add(25*time.Minute))),
				fmt.Sprintf("Deploying %v branch5", ts(releaseTime.Add(30*time.Minute))), // 10 + 5 + 10 + 5
				fmt.Sprintf("%v Triggered Successfully", ts(releaseTime.Add(30*time.Minute))),
			},
			expectedFinalTime: releaseTime.Add(30 * time.Minute),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cliMock := &mock.CLI{}
			timeMock := mock.NewMockedTime(releaseTime)
			deploy := &Deploy{
				c:               cliMock,
				releaseBranches: test.branches,
				host:            &mock.RepositoryClient{},
				time:            timeMock,
			}
			intervalDurations, _, err := deploy.calculateReleaseTime(test.intervals, "1ms")
			if err != nil {
				t.Errorf("NewDeploy().Do() returned error: %v", err)
			}
			err = deploy.deployToAllReleaseBranches(context.Background(), intervalDurations, &ergo.Release{}, false)
			if err != nil {
				t.Errorf("Deploy.deployToAllReleaseBranches() returned error %v", err)
			}
			if len(test.expectedPrints) != len(cliMock.PrintLines) {
				t.Errorf("Expected %d number of prints to the mock, got %d", len(test.expectedPrints), len(cliMock.PrintLines))
			} else {
				for i, expected := range test.expectedPrints {
					got := cliMock.PrintLines[i]
					if !reflect.DeepEqual(expected, got) {
						t.Errorf("Expected print %d: '%v' to equal '%v'", i, expected, got)
					}
				}
			}
			if test.expectedFinalTime != timeMock.CurrentTime {
				t.Errorf("Expected final time to equal %v, got %v", test.expectedFinalTime, timeMock.CurrentTime)
			}
		})
	}
}
