package repo

import (
	"fmt"
	"strings"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

// Repo describes a git repository
type Repo struct {
	repoURL    string
	path       string
	remoteName string

	repo   *git.Repository
	remote *git.Remote
}

// DiffCommitBranch commits ahead and commits behind for a given branch and base branch
type DiffCommitBranch struct {
	Branch     string
	BaseBranch string
	Ahead      []*object.Commit
	Behind     []*object.Commit
}

// NewFromPath instantiates a Repo loading it from a directory on disk
func NewFromPath(path, remoteName string) (*Repo, error) {
	var err error

	r := &Repo{
		path:       path,
		remoteName: remoteName,
	}

	if r.path == "" {
		return nil, fmt.Errorf("no path provided")
	}

	r.repo, err = git.PlainOpen(r.path)
	if err != nil {
		return nil, err
	}

	err = r.setGitRemote()
	if err != nil {
		return nil, fmt.Errorf("error set")
	}

	return r, nil
}

// NewClone instantiates a Repo loading it from a directory on disk
func NewClone(repoURL, path, remoteName string) (*Repo, error) {
	var err error

	r := &Repo{
		repoURL:    repoURL,
		path:       path,
		remoteName: remoteName,
	}

	if r.path == "" {
		return nil, fmt.Errorf("no path to clone to")
	}
	if r.repoURL == "" {
		return nil, fmt.Errorf("no url to clone from")
	}

	r.repo, err = git.PlainClone(r.path, false, &git.CloneOptions{
		URL:               r.repoURL,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})
	if err != nil {
		return nil, err
	}

	err = r.setGitRemote()
	if err != nil {
		return nil, fmt.Errorf("error set")
	}

	return r, nil
}

// Fetch fetches from the default remote
func (r *Repo) Fetch() error {
	fmt.Printf("Fetching remote %s (use --skipFetch to skip)\n", r.remoteName)
	err := r.remote.Fetch(&git.FetchOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), "already up-to-date") {
			return fmt.Errorf("unable to fetch remote %s: %s", r.remoteName, err)
		}
		// simply a notice, not an error
		fmt.Println(err)
	}

	return nil
}

// GitRepo exposes a git.Repository
func (r *Repo) GitRepo() *git.Repository {
	return r.repo
}

// GitRemote exposes a git.Remote
func (r *Repo) GitRemote() *git.Remote {
	return r.remote
}

func (r *Repo) setGitRemote() error {
	remote, err := r.repo.Remote(r.remoteName)
	if err != nil {
		return fmt.Errorf("error loading remote %s:%v", r.remoteName, err)
	}

	r.remote = remote

	return nil
}

// OrganizationName default remote's organization or user
func (r *Repo) OrganizationName() string {
	parts := strings.Split(r.remote.Config().URLs[0], "/")
	name := parts[len(parts)-2]
	// if remote is set by ssh instead of https
	if strings.Contains(name, ":") {
		return name[strings.LastIndex(name, ":")+1:]
	}

	return name
}

// Name the name of the repo as a suffix of the clone url (excluding .git) of the default remote
func (r *Repo) Name() string {
	parts := strings.Split(r.remote.Config().URLs[0], "/")

	return strings.TrimSuffix(parts[len(parts)-1], ".git")
}

// Path the path where the repo resides
func (r *Repo) Path() string {
	return r.path
}
