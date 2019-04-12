package release

import (
	"context"

	"github.com/blang/semver"
	"github.com/hashicorp/go-version"
	"github.com/taxibeat/ergo"
)

// Version is responsible to describe the actions of visioning.
type Version struct {
	host       ergo.Host
	baseBranch string
}

// NewVersion initializes and return a new Version object.
func NewVersion(host ergo.Host, baseBranch string) *Version {
	return &Version{host: host, baseBranch: baseBranch}
}

// NextVersion finds the next version according to major/minor/patch pattern.
func (v Version) NextVersion(ctx context.Context, inputVersion, suffix string, major, minor bool) (*ergo.Version, error) {
	baseBranchReference, err := v.host.GetRef(ctx, v.baseBranch)
	if err != nil {
		return nil, err
	}

	// Check for force version.
	if forceVersion := forceVersion(inputVersion, suffix); forceVersion != "" {
		return &ergo.Version{Name: forceVersion, SHA: baseBranchReference.SHA}, nil
	}

	// Calculate the new version name according to remote tags names.
	newVersion := semver.Version{}

	lastRelease, err := v.host.LastRelease(ctx)
	if err != nil {
		return nil, err
	}

	prevVersion, err := v.getVersionFromLastRelease(lastRelease)
	if err != nil {
		return nil, err
	}

	newVersion = increaseVersion(prevVersion, major, minor)

	if suffix == "" {
		return &ergo.Version{Name: newVersion.String(), SHA: baseBranchReference.SHA}, nil
	}

	newVersion, err = v.addSuffix(ctx, baseBranchReference.SHA, suffix, newVersion, prevVersion)
	if err != nil {
		return nil, err
	}

	return &ergo.Version{Name: newVersion.String(), SHA: baseBranchReference.SHA}, nil

}

// getVersionFromLastRelease gets the version from last release object.
func (v Version) getVersionFromLastRelease(lastRelease *ergo.Release) (semver.Version, error) {
	if lastRelease == nil {
		return semver.Make("0.0.0")
	}
	return v.parseLastVersion(lastRelease.TagName)
}

// parseLastVersion parses the version of the latest release and returns the semver.Version object.
func (v Version) parseLastVersion(tagName string) (semver.Version, error) {
	var tempVersion semver.Version

	semVersion, err := version.NewSemver(tagName)

	// If sem version does not have the proper format then create a default version.
	if err != nil {
		tempVersion, err = semver.Make("0.0.0")
		if err != nil {
			return semver.Version{}, err
		}
		return tempVersion, nil
	}

	tempVersion, err = semver.Make(semVersion.String())
	if err != nil {
		return semver.Version{}, nil
	}

	return semver.ParseTolerant(tempVersion.String())
}

// forceVersion returns the input version with the suffix if the parameters present otherwise return empty string.
func forceVersion(inputVersion, suffix string) string {
	if inputVersion == "" {
		return ""
	}
	if suffix != "" {
		return inputVersion + "-" + suffix
	}
	return inputVersion
}

// increase increases the version according to given flag (major or minor or nothing/patch)
func increaseVersion(prevVersion semver.Version, major, minor bool) semver.Version {
	if major {
		return semver.Version{Major: prevVersion.Major + 1, Minor: 0, Patch: 0}
	}
	if minor {
		return semver.Version{Major: prevVersion.Major, Minor: prevVersion.Minor + 1, Patch: 0}
	}
	return semver.Version{Major: prevVersion.Major, Minor: prevVersion.Minor, Patch: prevVersion.Patch + 1}
}

// addSuffix if suffix version is present add it.
func (v Version) addSuffix(
	ctx context.Context,
	latestCommitSHA, suffix string,
	newVersion, prevVersion semver.Version,
) (semver.Version, error) {
	refFromTag, err := v.host.GetRefFromTag(ctx, prevVersion.String())
	if err != nil {
		return semver.Version{}, err
	}

	if latestCommitSHA == refFromTag.SHA {
		newVersion = prevVersion
	}

	// Add the suffix to the end.
	newVersion.Pre = append(newVersion.Pre, semver.PRVersion{VersionStr: suffix})

	return newVersion, nil
}
