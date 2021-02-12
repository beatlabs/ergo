package release

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/beatlabs/ergo"
	"github.com/beatlabs/ergo/cli"
	"github.com/pkg/errors"
)

// Deploy is responsible to describe the release process.
type Deploy struct {
	c                   ergo.CLI
	host                ergo.Host
	baseBranch          string
	releaseBodyFind     string
	releaseBodyReplace  string
	releaseBranches     []string
	releaseBodyBranches map[string]string
}

// NewDeploy initialize and return a new Deploy object.
func NewDeploy(
	c ergo.CLI,
	host ergo.Host,
	baseBranch, releaseBodyFind, releaseBodyReplace string,
	releaseBranches []string,
	releaseBodyBranches map[string]string,
) *Deploy {
	return &Deploy{
		c:                   c,
		host:                host,
		baseBranch:          baseBranch,
		releaseBodyFind:     releaseBodyFind,
		releaseBodyReplace:  releaseBodyReplace,
		releaseBranches:     releaseBranches,
		releaseBodyBranches: releaseBodyBranches,
	}
}

// Do is responsible for deploying the latest release.
func (r *Deploy) Do(ctx context.Context, releaseIntervalInput, releaseOffsetInput string, allowForcePush bool) error {
	release, err := r.host.LastRelease(ctx)
	if err != nil {
		return err
	}

	r.c.PrintColorizedLine("REPO: ", r.host.GetRepoName(), cli.WarningType)
	r.c.PrintLine("Deploying ", release.ReleaseURL)
	r.c.PrintLine("Deployment start times are estimates.")

	intervalDuration, releaseTimer, err := r.calculateReleaseTime(releaseIntervalInput, releaseOffsetInput)
	if err != nil {
		return err
	}

	releaseTime := *releaseTimer

	r.printReleaseTimeBoard(releaseTime, r.releaseBranches, intervalDuration)

	confirm, err := r.c.Confirmation("Deployment", "No deployment", "")
	if err != nil {
		return err
	}
	if !confirm {
		return nil
	}

	if releaseTime.Before(time.Now()) {
		return errors.New("deployment stopped since first released time has passed. Please run again")
	}

	untilReleaseTime := time.Until(releaseTime)
	r.c.PrintLine("Deployment will start in", untilReleaseTime.String())
	time.Sleep(untilReleaseTime)

	for i, branch := range r.releaseBranches {
		if i != 0 {
			time.Sleep(intervalDuration)
			releaseTime = releaseTime.Add(intervalDuration)
		}
		r.c.PrintLine("Deploying", time.Now().Format("15:04:05"), branch)

		if errRelease := r.host.UpdateBranchFromTag(ctx, release.TagName, branch, allowForcePush); errRelease != nil {
			return errRelease
		}
		r.c.PrintLine(time.Now().Format("15:04:05"), "Triggered Successfully")

		err = r.updateHostReleaseBody(ctx, r.releaseBodyBranches, branch, r.releaseBodyFind, r.releaseBodyReplace)
		if err != nil {
			return err
		}
	}

	return nil
}

// calculateReleaseTime calculate from string the interval between the releases.
func (r *Deploy) calculateReleaseTime(releaseInterval, releaseOffset string) (time.Duration, *time.Time, error) {
	intervalDuration, err := time.ParseDuration(releaseInterval)
	if err != nil {
		return 0, nil, errors.Wrap(err, "error parsing interval")
	}
	offsetDuration, err := time.ParseDuration(releaseOffset)
	if err != nil {
		return 0, nil, errors.Wrap(err, "error parsing duration")
	}
	releaseTime := time.Now().Add(offsetDuration)
	return intervalDuration, &releaseTime, nil
}

// printReleaseTimeBoard print the release time board.
func (r *Deploy) printReleaseTimeBoard(releaseTime time.Time, releaseBranches []string, intervalDuration time.Duration) {
	var times [][]string

	for _, branch := range releaseBranches {
		timesRow := []string{branch, releaseTime.Format("15:04 MST")}
		releaseTime = releaseTime.Add(intervalDuration)
		times = append(times, timesRow)
	}

	headers := []string{"Branch", "Start Time"}
	cli.NewCLI().PrintTable(headers, times)
}

// updateHostReleaseBody update the host release body.
func (r *Deploy) updateHostReleaseBody(ctx context.Context, branchMap map[string]string, branch, suffixFind, suffixReplace string) error {
	branchText, ok := branchMap[branch]
	if !ok {
		branchText = branch
	}
	if suffixFind != "" {
		err := r.updateReleaseBodySuffix(ctx, branchText, suffixFind, suffixReplace)
		if err != nil {
			return err
		}
	}
	return nil
}

// updateReleaseBodySuffix update the release body suffixes.
func (r *Deploy) updateReleaseBodySuffix(ctx context.Context, branchText, suffixFind, suffixReplace string) error {
	t := time.Now()
	release, err := r.host.LastRelease(ctx)
	if err != nil {
		return err
	}

	findText := fmt.Sprintf("%s ![](https://img.shields.io/badge/released%s)", branchText, suffixFind)
	replaceText := fmt.Sprintf("%s ![](https://img.shields.io/badge/released-%d_%s_%d_%02d:%02d%s)",
		branchText, t.Day(), t.Month(), t.Year(), t.Hour(), t.Minute(), suffixReplace)
	newBody := strings.Replace(release.Body, findText, replaceText, -1)
	release.Body = newBody
	_, err = r.host.EditRelease(ctx, release)
	if err != nil {
		return err
	}
	return nil
}
