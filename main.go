package main

import (
	"flag"
	"fmt"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/rodaine/table"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

// DiffCommitBranch commits ahead and commits behind for a given branch and base branch
type DiffCommitBranch struct {
	branch     string
	baseBranch string
	ahead      []*object.Commit
	behind     []*object.Commit
}

func main() {
	var repoURL string
	var directory string
	var skipFetch bool
	var baseBranch string
	var branchesString string
	var err error
	var diff []DiffCommitBranch

	flag.StringVar(&repoURL, "repoUrl", "", "git repo Url. ssh and https supported")
	flag.StringVar(&directory, "directory", "", "Location to store or retrieve from the repo")
	flag.BoolVar(&skipFetch, "skipFetch", false, "When true on an existing repo, it will not fetch the latest commits from remote")
	flag.StringVar(&baseBranch, "baseBranch", "master", "Base Reference for comparison. If empty, tags will be used")
	flag.StringVar(&branchesString, "branches", "", "Comma separated list of branches")
	flag.Parse()

	repo, err := loadOrClone(repoURL, directory, "origin", skipFetch)
	if err != nil {
		fmt.Printf("Error loading repo:%s\n", err)
		return
	}

	if branchesString == "" {
		fmt.Printf("no branches to compare, use -branches\n")
		return
	}

	branches := strings.Split(branchesString, ",")
	for _, branch := range branches {
		ahead, behind, err := compareBranch(repo, baseBranch, branch, directory)
		if err != nil {
			fmt.Printf("error comparing %s %s:%s\n", baseBranch, branch, err)
			return
		}
		branchCommitDiff := DiffCommitBranch{
			branch:     branch,
			baseBranch: baseBranch,
			ahead:      ahead,
			behind:     behind,
		}
		diff = append(diff, branchCommitDiff)
	}

	printBranchCompare(diff)
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

func printBranchCompare(commitDiffBranches []DiffCommitBranch) {
	blue := color.New(color.FgCyan)
	yellow := color.New(color.FgYellow)
	fmt.Println()
	blue.Print("BASE: ")
	yellow.Println(commitDiffBranches[0].baseBranch)

	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("Branch", "Behind", "Ahead")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for _, diffBranch := range commitDiffBranches {
		tbl.AddRow(diffBranch.branch, len(diffBranch.behind), len(diffBranch.ahead))
	}

	tbl.Print()
}
