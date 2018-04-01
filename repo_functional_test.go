package main

import (
	"os"
	"testing"
)

func TestLoadOrClone(t *testing.T) {
	directory := "/tmp/test-ergo-repo"
	repoURL := "https://github.com/dbaltas/ergo.git"
	skipFetch := true

	repoFromClone, err := loadOrClone(repoURL, directory, "origin", skipFetch)
	if err != nil || repoFromClone == nil {
		t.Errorf("Error cloning repo:%s\n", err)
		return
	}
	repoFromSecondClone, err := loadOrClone(repoURL, directory, "origin", skipFetch)
	if repoFromSecondClone != nil {
		t.Errorf("Expected 'repository already exists' error")
		return
	}
	repoFromPath, err := loadOrClone("", directory, "origin", skipFetch)
	if err != nil || repoFromPath == nil {
		t.Errorf("error loading repo from path: %s", err)
		return
	}

	err = os.RemoveAll(directory)

	if err != nil {
		t.Errorf("error cleaning up %s: %s", directory, err)
	}
	// if branchesString == "" {
	// 	fmt.Printf("no branches to compare, use -branches\n")
	// 	return
	// }

	// branches := strings.Split(branchesString, ",")
	// for _, branch := range branches {
	// 	ahead, behind, err := compareBranch(repo, baseBranch, branch, directory)
	// 	if err != nil {
	// 		fmt.Printf("error comparing %s %s:%s\n", baseBranch, branch, err)
	// 		return
	// 	}
	// 	branchCommitDiff := DiffCommitBranch{
	// 		branch:     branch,
	// 		baseBranch: baseBranch,
	// 		ahead:      ahead,
	// 		behind:     behind,
	// 	}
	// 	diff = append(diff, branchCommitDiff)
	// }
}
