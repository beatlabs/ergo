package viper

import (
	"fmt"
	"testing"

	"github.com/spf13/viper"
)

func TestGetConfigShouldNotReturnError(t *testing.T) {
	v := viper.GetViper()
	v.Set("generic.base-branch", "opsss")
	v.Set("github.access-token", "abcd")
	v.Set("release.branch-map", map[string]string{"release-gr": ":greece:", "releases/greece": ":greece:"})
	v.Set("github.release-body-prefix", "Changelog:")
	v.Set("release.on-deploy.body-branch-suffix-find", "-No-red.svg")
	v.Set("release.on-deploy.body-branch-suffix-replace", "-green.svg")
	v.Set("github.default-owner", "acme")
	v.Set("github.default-repo", "my-repo")

	vipOpts := NewOptions()
	vipOpts.AccToken = "abcd"
	vipOpts.RefreshConfig()
	_, err := vipOpts.GetConfig()

	if err != nil {
		t.Error("expected get config not to return error.")
	}
}

func TestGetConfigShouldReturnError(t *testing.T) {
	v := viper.GetViper()
	v.Set("github.access-token", "abcd")
	v.Set("release.branch-map", map[string]string{"release-gr": ":greece:", "releases/greece": ":greece:"})
	v.Set("github.release-body-prefix", "Changelog:")
	v.Set("release.on-deploy.body-branch-suffix-find", "-No-red.svg")
	v.Set("release.on-deploy.body-branch-suffix-replace", "-green.svg")
	v.Set("github.default-owner", "acme")
	v.Set("github.default-repo", "my-repo")

	vipOpts := NewOptions()
	vipOpts.RefreshConfig()
	_, err := vipOpts.GetConfig()

	if err == nil {
		t.Error("expected get config to return error.")
	}
}

func TestGetConfigShouldOverwriteEmptyValuesAfterRefresh(t *testing.T) {
	v := viper.GetViper()
	v.Set("github.access-token", "abcd")
	v.Set("release.branch-map", map[string]string{"release-gr": ":greece:", "releases/greece": ":greece:"})
	v.Set("github.release-body-prefix", "Changelog:")
	v.Set("release.on-deploy.body-branch-suffix-find", "-No-red.svg")
	v.Set("release.on-deploy.body-branch-suffix-replace", "-green.svg")
	v.Set("github.default-owner", "acme")
	v.Set("github.default-repo", "my-repo")
	v.Set("generic.base-branch", "staging-develop")

	vipOpts := NewOptions()
	vipOpts.AccToken = "abcd"
	vipOpts.RefreshConfig()
	opts, _ := vipOpts.GetConfig()

	if opts.RepoName != "my-repo" || opts.Organization != "acme" || opts.BaseBranch != "staging-develop" {
		t.Error("expected refresh config to overwrite empty.")
	}
}

func TestGetConfigShouldOverwriteStatusBranchesAfterRefresh(t *testing.T) {
	v := viper.GetViper()
	v.Set("github.access-token", "abcd")
	v.Set("github.default-owner", "acme")
	v.Set("github.default-repo", "my-repo")
	v.Set("generic.status-branches", "develop,master,staging")

	vipOpts := NewOptions()
	vipOpts.AccToken = "abcd"
	vipOpts.RefreshConfig()
	opts, _ := vipOpts.GetConfig()

	if opts.Branches[0] != "develop" && opts.Branches[1] != "master" && opts.Branches[1] != "staging" {
		t.Error("expected refresh config to overwrite generic status branches.")
	}

	vipOpts.RefreshConfig()
	opts, _ = vipOpts.GetConfig()

	v.Set("repos.my-repo.status-branches", "develop,stable")

	if opts.Branches[0] != "develop" && opts.Branches[1] != "stable" {
		t.Error("expected refresh config to overwrite repo status branches.")
	}
}

func TestGetConfigShouldOverwriteReleaseBranchesFromSpecificKeyAfterRefresh(t *testing.T) {
	v := viper.GetViper()
	v.Set("github.access-token", "abcd")
	v.Set("github.default-owner", "acme")
	v.Set("github.default-repo", "my-repo")
	v.Set("generic.release-branches", "release-cn,release-vn")

	vipOpts := NewOptions()
	vipOpts.AccToken = "abcd"
	vipOpts.RefreshConfig()
	opts, _ := vipOpts.GetConfig()

	if opts.ReleaseBranches[0] != "release-cn" && opts.ReleaseBranches[1] != "release-vn" {
		t.Error("expected refresh config to overwrite generic release branches.")
	}
}

func TestGetConfigShouldOverwriteReleaseBranchesFromGenericKeyAfterRefresh(t *testing.T) {
	v := viper.GetViper()
	v.Set("github.access-token", "abcd")
	v.Set("github.default-owner", "acme")
	v.Set("github.default-repo", "my-repo")
	v.Set("generic.release-branches", "release-cn,release-vn")
	v.Set("repos.my-repo.release-branches", "china,vietnam")

	vipOpts := NewOptions()
	vipOpts.AccToken = "abcd"
	vipOpts.RefreshConfig()
	opts, _ := vipOpts.GetConfig()

	fmt.Println(opts.ReleaseBranches)
	if opts.ReleaseBranches[0] != "china" && opts.ReleaseBranches[1] != "vietnam" {
		t.Error("expected refresh config to overwrite repo release branches.")
	}
}

func TestGetConfigShouldNotOverwriteNonEmptyValuesAfterRefresh(t *testing.T) {
	vipOpts := NewOptions()
	vipOpts.AccToken = "abcd"
	vipOpts.Organization = "someorg"
	vipOpts.RepoName = "somerepo"

	v := viper.GetViper()
	v.Set("github.access-token", "abcd")
	v.Set("release.branch-map", map[string]string{"release-gr": ":greece:", "releases/greece": ":greece:"})
	v.Set("github.release-body-prefix", "Changelog:")
	v.Set("release.on-deploy.body-branch-suffix-find", "-No-red.svg")
	v.Set("release.on-deploy.body-branch-suffix-replace", "-green.svg")
	v.Set("github.default-owner", "acme")
	v.Set("github.default-repo", "my-repo")

	vipOpts.RefreshConfig()
	opts, _ := vipOpts.GetConfig()

	if opts.RepoName != "somerepo" || opts.Organization != "someorg" {
		t.Error("expected refresh config should not overwrite non empty.")
	}
}
