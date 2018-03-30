package main

import (
	"flag"
	"fmt"
	"os/exec"
	"strings"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

var baseSHA1 string

func main() {
	var repoURL string
	var directory string
	var skipFetch bool
	var goBackTags int
	var err error
	flag.StringVar(&repoURL, "repoUrl", "", "git repo Url. ssh and https supported")
	flag.StringVar(&directory, "directory", "", "Location to store or retrieve from the repo")
	flag.BoolVar(&skipFetch, "skipFetch", false, "When true on an existing repo, it will not fetch the latest commits from remote")
	flag.IntVar(&goBackTags, "goBackTags", 1, "Number of tags to go back for base. Defaults to 1, last tag")
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

	gitFindTagSha1Cmd := fmt.Sprintf("git log --tags --no-walk --pretty=\"format:%%d\" | sed %dq | sed 's/[()]//g' | sed s/,[^,]*$// | sed  's ......  '", goBackTags)
	cmd := fmt.Sprintf("cd %s && %s |tail -n 1", directory, gitFindTagSha1Cmd)
	out, err := exec.Command("sh", "-c", cmd).Output()

	if err != nil {
		fmt.Printf("error executing %s:%s\n", cmd, err)
		return
	}

	baseTag := strings.TrimSpace(string(out))

	reference := fmt.Sprintf("refs/tags/%s", baseTag)
	baseRef, err := repo.Reference(plumbing.ReferenceName(reference), true)
	if err != nil {
		fmt.Printf("Could not load tag %s:%s", baseTag, err)
		return
	}

	baseSHA1 = baseRef.Hash().String()

	fmt.Printf("\n********** B A S E **********\ntag: %s\nsha1: %s\n*****************************\n", baseTag, baseSHA1)

	displayBranch(repo, "develop")
	displayBranch(repo, "master")
	displayBranch(repo, "release-gr")
	displayBranch(repo, "release-pe")
	displayBranch(repo, "release-cl")
	displayBranch(repo, "release-co")
}

func displayBranch(repo *git.Repository, branch string) {
	fmt.Printf("\n%s\n", branch)
	var reference string
	reference = fmt.Sprintf("refs/remotes/origin/%s", branch)

	releaseRef, err := repo.Reference(plumbing.ReferenceName(reference), true)
	if err != nil {
		fmt.Printf("Error loading reference %s:%s", reference, err)
		return
	}

	cIter, err := repo.Log(&git.LogOptions{From: releaseRef.Hash()})

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

		if commit.Hash.String() == baseSHA1 {
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
