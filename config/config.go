package config

import "gopkg.in/src-d/go-git.v4/plumbing/format/config"

// Options include the base configuration for creating draft, releasing deployments, statusing etc.
type Options struct {
	BaseBranch      string
	Branches        []string
	ReleaseBranches []string
	AccToken        string

	ReleaseBodyBranches map[string]string
	ReleaseBodyPrefix   string
	ReleaseBodyFind     string
	ReleaseBodyReplace  string

	GenericRemote string

	Organization string
	RepoName     string
}

// Config interface describes the config initialization.
type Config interface {
	InitConfig() error
	RefreshConfig()
	GetConfig() (*config.Options, error)
}
