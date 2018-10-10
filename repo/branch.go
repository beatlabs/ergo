package repo

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

// CurrentBranch returns the currently checked out branch
func (r *Repo) CurrentBranch() (string, error) {
	cmd := fmt.Sprintf("cd %s && git rev-parse --abbrev-ref HEAD", r.path)
	out, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return "", errors.Wrap(err, "executing external command")
	}

	return strings.TrimSpace(string(out)), nil
}

// CompareBranch lists the commits ahead and behind of a targetBranch compared
// to a baseBranch.
func (r *Repo) CompareBranch(baseBranch, branch string) ([]*object.Commit, []*object.Commit, error) {
	commonAncestor, err := mergeBase(baseBranch, branch, r.path)
	if err != nil {
		return nil, nil, errors.Wrap(err, "executing merge-base")
	}

	ahead, err := commitsAhead(r.repo, branch, commonAncestor)
	if err != nil {
		return nil, nil, errors.Wrap(err, "comparing branches")
	}
	behind, err := commitsAhead(r.repo, baseBranch, commonAncestor)
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
