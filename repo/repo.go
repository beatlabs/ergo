package repo

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

// Repo describes a git repository
type Repo struct {
	repoURL    string
	directory  string
	remoteName string

	Repo *git.Repository
}

// DiffCommitBranch commits ahead and commits behind for a given branch and base branch
type DiffCommitBranch struct {
	Branch     string
	BaseBranch string
	Ahead      []*object.Commit
	Behind     []*object.Commit
}

// New instantiates a Repo
func New(repoURL, directory, remoteName string) *Repo {
	return &Repo{
		repoURL:    repoURL,
		directory:  directory,
		remoteName: remoteName,
	}
}

// LoadOrClone clones a repo in a given directory. Or loads a repo if no repoUrl is provided
func (r *Repo) LoadOrClone(skipFetch bool) (*git.Repository, error) {
	var repo *git.Repository
	var err error

	if r.directory == "" {
		return nil, fmt.Errorf("no directory provided")
	}

	if r.repoURL != "" {
		repo, err = git.PlainClone(r.directory, false, &git.CloneOptions{
			URL:               r.repoURL,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		})

		if err != nil {
			return nil, err
		}
	}

	if repo == nil {
		repo, err = git.PlainOpen(r.directory)
		if err != nil {
			return nil, err
		}
	}

	remote, err := repo.Remote(r.remoteName)
	if err != nil {
		fmt.Printf("error loading remote %s:%s", r.remoteName, err)
		return repo, err
	}

	if !skipFetch {
		fmt.Printf("Fetching remote %s (use --skipFetch to skip)\n", r.remoteName)
		err = remote.Fetch(&git.FetchOptions{})
		if err != nil {
			if !strings.Contains(err.Error(), "already up-to-date") {
				return repo, fmt.Errorf("unable to fetch remote %s: %s", r.remoteName, err)
			}
			// simply a notice, not an error
			fmt.Println(err)
		}
	}
	r.Repo = repo

	return repo, nil
}

// CurrentBranch returns the currently checked out branch
func (r *Repo) CurrentBranch() (string, error) {
	cmd := fmt.Sprintf("cd %s && git rev-parse --abbrev-ref HEAD", r.directory)
	out, err := exec.Command("sh", "-c", cmd).Output()

	if err != nil {
		return "", errors.Wrap(err, "executing external command")
	}

	return strings.TrimSpace(string(out)), nil
}

// CompareBranch lists the commits ahead and behind of a targetBranch compared
// to a baseBranch.
func (r *Repo) CompareBranch(baseBranch, branch string) ([]*object.Commit, []*object.Commit, error) {
	commonAncestor, err := mergeBase(baseBranch, branch, r.directory)
	if err != nil {
		return nil, nil, errors.Wrap(err, "executing merge-base")
	}

	ahead, err := commitsAhead(r.Repo, branch, commonAncestor)
	if err != nil {
		return nil, nil, errors.Wrap(err, "comparing branches")
	}
	behind, err := commitsAhead(r.Repo, baseBranch, commonAncestor)
	if err != nil {
		return nil, nil, errors.Wrap(err, "comparing branches")
	}

	return ahead, behind, nil
}

func commitsAhead(repo *git.Repository, branch string, commonAncestor string) ([]*object.Commit, error) {
	reference := fmt.Sprintf("refs/remotes/origin/%s", branch)
	ref, err := repo.Reference(plumbing.ReferenceName(reference), true)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("loading reference %s", reference))
	}

	cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, errors.Wrap(err, "branch log")
	}
	defer cIter.Close()

	var ahead []*object.Commit
	for {
		commit, err := cIter.Next()
		if err != nil {
			return nil, errors.Wrap(err, "iterating commits")
		}

		if commit.Hash.String() == commonAncestor {
			break
		}
		ahead = append(ahead, commit)
	}

	return ahead, nil
}

func mergeBase(branch1 string, branch2 string, directory string) (string, error) {
	cmd := fmt.Sprintf("cd %s && git merge-base origin/%s origin/%s", directory, branch1, branch2)
	out, err := exec.Command("sh", "-c", cmd).Output()

	if err != nil {
		return "", errors.Wrap(err, "executing external command")
	}

	return strings.TrimSpace(string(out)), nil
}

// FormatMessage formats the commit's message
func FormatMessage(c *object.Commit, firstLinePrefix string, nextLinesPrefix string, lineSeparator string) string {
	outputStrings := []string{}
	maxLines := 6

	lines := strings.Split(c.Message, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		prefix := ""

		if len(outputStrings) == 0 {
			prefix = firstLinePrefix
		} else {
			prefix = nextLinesPrefix
		}

		outputStrings = append(outputStrings, fmt.Sprintf("%s%s", prefix, strings.TrimSpace(line)))

		if len(outputStrings) >= maxLines {
			break
		}
	}
	return strings.Join(outputStrings, lineSeparator)
}
