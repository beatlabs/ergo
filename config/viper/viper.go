package viper

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"github.com/thebeatapp/ergo/config"
)

// Options struct.
type Options struct {
	config.Options

	branchesString        string
	releaseBranchesString string
}

// NewOptions factory.
func NewOptions() *Options {
	return &Options{}
}

// InitConfig initializes configuration.
func (o *Options) InitConfig() (*config.Options, error) {
	err := initConfigFromFile()
	if err != nil {
		return nil, err
	}

	o.AccToken = viper.GetString("github.access-token")

	o.ReleaseBodyBranches = viper.GetStringMapString("release.branch-map")
	o.ReleaseBodyPrefix = viper.GetString("github.release-body-prefix")
	o.ReleaseBodyFind = viper.GetString("release.on-deploy.body-branch-suffix-find")
	o.ReleaseBodyReplace = viper.GetString("release.on-deploy.body-branch-suffix-replace")

	o.Organization = viper.GetString("github.default-owner")
	o.RepoName = viper.GetString("github.default-repo")

	return &o.Options, nil
}

// RefreshConfig refreshes configuration.
func (o *Options) RefreshConfig() {
	o.setGenericConfigs()
	o.setStatusBranchConfig()
	o.setReleaseBranchesConfig()
}

// GetConfig gets configuration.
func (o *Options) GetConfig() (*config.Options, error) {
	field, valid := o.validateOptions()
	if !valid {
		return nil, errors.New("invalid field: " + field)
	}

	return &o.Options, nil
}

// SetBranchesString overwrites the branches string (comma delimited).
func (o *Options) SetBranchesString(branchesString string) {
	o.branchesString = branchesString
}

// SetReleaseBranchesString sets the release branches.
func (o *Options) SetReleaseBranchesString(releaseBranchesString string) {
	o.releaseBranchesString = releaseBranchesString
}

// SetReleaseBranches set the release branches.
func (o *Options) SetReleaseBranches(releaseBranchesString string) {
	o.ReleaseBranches = strings.Split(releaseBranchesString, ",")
}

// setGenericConfigs sets the generic configs.
func (o *Options) setGenericConfigs() {
	if o.BaseBranch == "" {
		o.BaseBranch = viper.GetString("generic.base-branch")
	}
	if o.Organization == "" {
		o.Organization = viper.GetString("github.default-owner")
	}
	if o.RepoName == "" {
		o.RepoName = viper.GetString("github.default-repo")
	}
}

// setStatusBranchConfig sets the status branch config.
func (o *Options) setStatusBranchConfig() {
	if o.branchesString == "" && o.RepoName != "" {
		o.branchesString = viper.GetString(fmt.Sprintf("repos.%s.status-branches", o.RepoName))
	}
	if o.branchesString == "" {
		o.branchesString = viper.GetString("generic.status-branches")
	}
	if o.branchesString != "" {
		o.Branches = strings.Split(o.branchesString, ",")
	}
}

// setReleaseBranchesConfig sets the release branches config.
func (o *Options) setReleaseBranchesConfig() {
	if o.releaseBranchesString == "" && o.RepoName != "" {
		o.releaseBranchesString = viper.GetString(fmt.Sprintf("repos.%s.release-branches", o.RepoName))
	}
	if o.releaseBranchesString == "" {
		o.releaseBranchesString = viper.GetString("generic.release-branches")
	}
	if o.releaseBranchesString != "" {
		o.ReleaseBranches = strings.Split(o.releaseBranchesString, ",")
	}
}

// initConfigFromFile initialize the config from yaml file.
func initConfigFromFile() error {
	home, err := homedir.Dir()
	if err != nil {
		return err
	}

	viper.AddConfigPath(home)
	viper.SetConfigName(".ergo")

	return viper.ReadInConfig()
}

// validateOptions validate the mandatory options.
func (o *Options) validateOptions() (string, bool) {
	if o.AccToken == "" {
		return "access token", false
	}

	if o.Organization == "" {
		return "organization", false
	}

	if o.RepoName == "" {
		return "repository", false
	}

	if o.BaseBranch == "" {
		return "base branch", false
	}

	return "", true
}
