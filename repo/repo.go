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

// DiffCommitBranch commits ahead and commits behind for a given branch and base branch
type DiffCommitBranch struct {
	Branch     string
	BaseBranch string
	Ahead      []*object.Commit
	Behind     []*object.Commit
}

// LoadOrClone clones a repo in a given directory. Or loads a repo if no repoUrl is provided
func LoadOrClone(repoURL string, directory string, remoteName string, skipFetch bool) (*git.Repository, error) {
	var repo *git.Repository
	var err error

	if directory == "" {
		return nil, fmt.Errorf("no directory provided")
	}

	if repoURL != "" {
		repo, err = git.PlainClone(directory, false, &git.CloneOptions{
			URL:               repoURL,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		})

		if err != nil {
			return nil, err
		}
	}

	if repo == nil {
		repo, err = git.PlainOpen(directory)
		if err != nil {
			return nil, err
		}
	}

	remote, err := repo.Remote(remoteName)
	if err != nil {
		fmt.Printf("error loading remote %s:%s", remoteName, err)
		return repo, err
	}

	if !skipFetch {
		fmt.Printf("Fetching remote %s (use -skipFetch to skip)\n", remoteName)
		err = remote.Fetch(&git.FetchOptions{})
		if err != nil {
			if !strings.Contains(err.Error(), "already up-to-date") {
				return repo, fmt.Errorf("unable to fetch remote %s: %s", remoteName, err)
			}
			fmt.Println(err)
		}
	}

	return repo, nil
}

func baseReference(repo *git.Repository, directory string, baseBranch string) (*plumbing.Reference, error) {
	baseRefText := fmt.Sprintf("refs/remotes/origin/%s", baseBranch)
	baseRef, err := repo.Reference(plumbing.ReferenceName(baseRefText), true)

	if err != nil {
		return nil, fmt.Errorf("could not load ref %s:%s", baseRefText, err)
	}

	return baseRef, nil
}

// CompareBranch lists the commits ahead and behind of a targetBranch compared to a baseBranch
func CompareBranch(repo *git.Repository, baseBranch string, branch string, directory string) ([]*object.Commit, []*object.Commit, error) {
	var behind []*object.Commit
	var ahead []*object.Commit

	commonAncestor, err := mergeBase(baseBranch, branch, directory)

	ahead, err = commitsAhead(repo, branch, commonAncestor)
	if err != nil {
		return nil, nil, errors.Wrap(err, "comparing branches")
	}
	behind, err = commitsAhead(repo, baseBranch, commonAncestor)
	if err != nil {
		return nil, nil, errors.Wrap(err, "comparing branches")
	}

	return ahead, behind, nil
}

func commitsAhead(repo *git.Repository, branch string, commonAncestor string) ([]*object.Commit, error) {
	var ahead []*object.Commit
	var reference string
	reference = fmt.Sprintf("refs/remotes/origin/%s", branch)
	ref, err := repo.Reference(plumbing.ReferenceName(reference), true)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("loading reference %s", reference))
	}

	cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, errors.Wrap(err, "branch log")
	}
	defer cIter.Close()

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
