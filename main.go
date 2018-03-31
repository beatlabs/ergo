package main

import (
	"os/exec"

	"gopkg.in/src-d/go-git.v4/plumbing/object"

	"github.com/pkg/errors"

	"flag"
	"fmt"
	"strings"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

var baseSHA1 string

func main() {
	var repoURL string
	var directory string
	var skipFetch bool
	var baseBranch string
	var branchesString string
	var err error
	flag.StringVar(&repoURL, "repoUrl", "", "git repo Url. ssh and https supported")
	flag.StringVar(&directory, "directory", "", "Location to store or retrieve from the repo")
	flag.BoolVar(&skipFetch, "skipFetch", false, "When true on an existing repo, it will not fetch the latest commits from remote")
	flag.StringVar(&baseBranch, "baseBranch", "", "Base Reference for comparison. If empty, tags will be used")
	flag.StringVar(&branchesString, "branches", "", "Comma separated list of branches")
	flag.Parse()

	repo, err := loadOrClone(repoURL, directory, "origin", skipFetch)
	if err != nil {
		fmt.Printf("Error loading repo:%s\n", err)
		return
	}

	baseRef, err := baseReference(repo, directory, baseBranch)
	if err != nil {
		fmt.Printf("Error loading reference:%s", err)
		return
	}

	baseSHA1 = baseRef.Hash().String()

	fmt.Printf("\nbase branch: %s\n", baseBranch)

	// compareBranch(repo, "develop")
	// compareBranch(repo, baseBranch, "master", directory)
	// compareBranch(repo, baseBranch, "release-gr", directory)
	fmt.Println(branchesString)
	branches := strings.Split(branchesString, ",")
	// branches := []string{"release-pe", "release-gr", "release-cl", "release-co"}
	for _, branch := range branches {
		ahead, behind, err := compareBranch(repo, baseBranch, branch, directory)
		if err != nil {
			fmt.Printf("error comparing %s %s:%s", baseBranch, branch, err)
			return
		}
		fmt.Println(branch)
		fmt.Println(len(ahead))
		fmt.Println(len(behind))
	}
	// compareBranch(repo, baseBranch, "release-cl", directory)
	// compareBranch(repo, baseBranch, "release-co", directory)
}

func loadOrClone(repoURL string, directory string, remoteName string, skipFetch bool) (*git.Repository, error) {
	var repo *git.Repository
	var err error

	if directory == "" {
		return nil, errors.New("no directory provided")
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
		fmt.Printf("Fetching remote %s (next time, use -skipFetch to skip)\n", remoteName)
		err = remote.Fetch(&git.FetchOptions{})
		if err != nil {
			if !strings.Contains(err.Error(), "already up-to-date") {
				msg := fmt.Sprintf("unable to fetch remote %s: %s\n", remoteName, err)
				return repo, errors.New(msg)
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
		msg := fmt.Sprintf("could not load ref %s:%s", baseRefText, err)
		return nil, errors.New(msg)
	}

	return baseRef, nil
}

func compareBranch(repo *git.Repository, baseBranch string, branch string, directory string) ([]*object.Commit, []*object.Commit, error) {
	var behind []*object.Commit
	var ahead []*object.Commit

	commonAncestor, err := mergeBase(baseBranch, branch, directory)

	ahead, err = commitsAhead(repo, branch, commonAncestor)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Comparing branches")
	}
	behind, err = commitsAhead(repo, baseBranch, commonAncestor)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Comparing branches")
	}

	return ahead, behind, nil
}

func commitsAhead(repo *git.Repository, branch string, commonAncestor string) ([]*object.Commit, error) {
	var ahead []*object.Commit
	var reference string
	reference = fmt.Sprintf("refs/remotes/origin/%s", branch)
	ref, err := repo.Reference(plumbing.ReferenceName(reference), true)
	if err != nil {
		return nil, errors.Wrap(err, "loading reference")
	}

	cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	defer cIter.Close()
	i := 0
	for {
		commit, err := cIter.Next()
		if err != nil {
			return nil, errors.Wrap(err, "iterating commits")
		}

		if commit.Hash.String() == commonAncestor {
			break
		}
		ahead = append(ahead, commit)

		// in case something went wrong, skip after 50 commits
		if i > 50 {
			return nil, errors.New("more than 50 commits difference. Write some code to bypass this")
		}
		i++
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
