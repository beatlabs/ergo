package main

import (
	"flag"
	"fmt"
	"os/exec"
	"strings"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

var lastReleased string

// Basic example of how to clone a repository using clone options.
func main() {
	var repoURL string
	var directory string
	var skipFetch bool
	var err error
	flag.StringVar(&repoURL, "repoUrl", "", "git repo Url. ssh and https supported")
	flag.StringVar(&directory, "directory", "", "Location to store or retrieve from the repo")
	flag.BoolVar(&skipFetch, "skipFetch", false, "When true on an existing repo, it will not fetch the latest commits from remote ")
	flag.Parse()

	if directory == "" {
		fmt.Println("no directory provided.")
		return
	}

	var repo *git.Repository
	if repoURL != "" {
		repo, err = git.PlainClone(directory, false, &git.CloneOptions{
			URL:               repoURL,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		})

		if err != nil {
			fmt.Printf("Error cloning repo '%s' to '%s':%s", repoURL, directory, err)
			return
		}
	}

	if repo == nil {
		repo, err = git.PlainOpen(directory)
		if err != nil {
			fmt.Printf("Error loading repo %s:%s", directory, err)
			return
		}
	}

	remoteName := "origin"
	remote, err := repo.Remote(remoteName)
	if err != nil {
		fmt.Printf("Error loading remote %s:%s", remoteName, err)
		return
	}

	if !skipFetch {
		fmt.Printf("Fetching remote %s (next time, use -skipFetch to skip)\n", remoteName)
		err = remote.Fetch(&git.FetchOptions{})
		if err != nil {
			if !strings.Contains(err.Error(), "already up-to-date") {
				fmt.Printf("Unable to fetch remote %s:%s\n", remoteName, err)
				return
			}
			// print already up to date
			fmt.Println(err)
		}
	}

	cmd := fmt.Sprintf("cd %s && git rev-list --tags --max-count=1", directory)
	out, err := exec.Command("sh", "-c", cmd).Output()

	if err != nil {
		fmt.Printf("error executing %s:%s\n", cmd, err)
		return
	}
	lastReleased = strings.TrimSpace(string(out))

	fmt.Printf("Last Released sha1: %s", lastReleased)

	// displayBranch(repo, "develop")
	displayBranch(repo, "master")
	displayBranch(repo, "release-gr")
	displayBranch(repo, "release-pe")
	displayBranch(repo, "release-cl")
	displayBranch(repo, "release-co")
}

func displayBranch(repo *git.Repository, branch string) {
	fmt.Printf("\n\n%s\n", branch)
	var reference string
	reference = fmt.Sprintf("refs/remotes/origin/%s", branch)

	releasePeRef, err := repo.Reference(plumbing.ReferenceName(reference), true)
	if err != nil {
		fmt.Printf("Error loading reference %s:%s", reference, err)
		return
	}

	cIter, err := repo.Log(&git.LogOptions{From: releasePeRef.Hash()})

	i := 0
	for {
		commit, _ := cIter.Next()
		i++
		var sha1 string
		var commitMessageFirstLine string
		message := commit.Message
		sha1 = commit.Hash.String()[:7]

		// ignore := false
		// if commit.Hash.String() == firstToInclude {
		// 	ignore = false
		// }

		// if ignore {
		// 	continue
		// }

		if commit.Hash.String() == lastReleased {
			break
		}

		lines := strings.Split(message, "\n")
		for _, line := range lines {
			if line != "" {
				commitMessageFirstLine = line
				break
			}
		}
		fmt.Printf("%s %s\n", sha1, commitMessageFirstLine)

		if i > 50 {
			cIter.Close()
			break
		}
	}

}
