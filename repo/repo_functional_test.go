package repo

import (
	"os"
	"testing"

	"gopkg.in/src-d/go-git.v4/plumbing/object"

	git "gopkg.in/src-d/go-git.v4"
)

func TestRepo(t *testing.T) {
	directory := "/tmp/ergo-functional-test-repo"
	repoURL := "https://github.com/dbaltas/ergo-functional-test-repo.git"
	skipFetch := true
	var repoFromClone, repoFromSecondClone, repoFromPath *git.Repository
	var err error
	var ahead, behind []*object.Commit

	// cleanup after test run
	defer func() {
		os.RemoveAll(directory)

		if err != nil {
			t.Errorf("error cleaning up %s: %v", directory, err)
			return
		}
	}()

	t.Run("Clone", func(t *testing.T) {
		repoFromClone, err = LoadOrClone(repoURL, directory, "origin", skipFetch)
		if err != nil || repoFromClone == nil {
			t.Errorf("Error cloning repo:%v\n", err)
			return
		}
	})

	t.Run("Clone already cloned", func(t *testing.T) {
		repoFromSecondClone, err = LoadOrClone(repoURL, directory, "origin", skipFetch)
		if repoFromSecondClone != nil {
			t.Errorf("Expected 'repository already exists' error")
			return
		}
	})

	t.Run("Load from disk already cloned", func(t *testing.T) {
		repoFromPath, err = LoadOrClone("", directory, "origin", skipFetch)
		if err != nil || repoFromPath == nil {
			t.Errorf("error loading repo from path: %v", err)
			return
		}
	})

	t.Run("Compare Branch", func(t *testing.T) {
		targetBranch := "ft-master"
		baseBranch := "ft-develop"
		ahead, behind, err = CompareBranch(repoFromPath, baseBranch, targetBranch, directory)
		if err != nil {
			t.Errorf("error comparing branches %s %s: %v", baseBranch, targetBranch, err)
			return
		}
		if len(ahead) != 0 {
			t.Errorf("expected %s to be 0 commits ahead of %s: actual:%d", targetBranch, baseBranch, len(ahead))
			return
		}
		if len(behind) != 2 {
			t.Errorf("expected %s to be 2 commits behind of %s: actual:%d", targetBranch, baseBranch, len(behind))
			return
		}
	})

	t.Run("Format Commit", func(t *testing.T) {
		commitMessage := FormatMessage(behind[0], "", "", "")

		if commitMessage != "feature-b" {
			t.Errorf("expected:%s, got:%s", "feature-b", commitMessage)
			return
		}
	})
}
