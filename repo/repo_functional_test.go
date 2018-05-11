package repo

import (
	"os"
	"testing"

	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

func TestRepo(t *testing.T) {
	var err error
	var ahead, behind []*object.Commit

	path := "/tmp/ergo-functional-test-repo"
	repoURL := "https://github.com/dbaltas/ergo-functional-test-repo.git"

	// r := New(repoURL, path, "origin")
	// cleanup after test run
	defer func() {
		os.RemoveAll(path)

		if err != nil {
			t.Errorf("error cleaning up %s: %v", path, err)
			return
		}
	}()

	r, err := NewClone(repoURL, path, "origin")

	if err != nil || r == nil {
		t.Errorf("Error cloning repo:%v\n", err)
		return
	}

	t.Run("Clone already cloned", func(t *testing.T) {
		_, err = NewClone(repoURL, path, "origin")

		if err == nil {
			t.Errorf("Expected 'repository already exists' error")
			return
		}
	})

	t.Run("Load from disk already cloned", func(t *testing.T) {
		r2, err := NewFromPath(path, "origin")
		if err != nil || r2 == nil {
			t.Errorf("error loading repo from path: %v", err)
			return
		}
	})

	t.Run("Compare Branch", func(t *testing.T) {
		targetBranch := "ft-master"
		baseBranch := "ft-develop"
		ahead, behind, err = r.CompareBranch(baseBranch, targetBranch)
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
