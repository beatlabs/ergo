package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"strings"

	"encoding/json"
	"github.com/taxibeat/ergo/config"
	"github.com/taxibeat/ergo/config/viper"
	"github.com/taxibeat/ergo/github"
	"github.com/taxibeat/ergo/release"
	"io/ioutil"
	"os"

	"context"
	"net/http"
)

type PushHook struct {
	Ref string `json:"ref"`
}

func githubHookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}


	jsonBody := getJsonBody(r)

	if !authorizeGithubHook(r, jsonBody) {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, "Unauthorized")

		return
	}

	vipOpts := viper.NewOptions()

	opts, err := vipOpts.InitConfig()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Could not initialize config")

		return
	}

	vipOpts.SetOrganization(os.Args[1])
	vipOpts.SetRepoName(os.Args[2])
	vipOpts.RefreshConfig()

	opts, err = vipOpts.GetConfig()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Could not initialize config")

		return
	}

	ph, err := parseHookFromReq(r, jsonBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Bad request")

		return
	}

	err = deleteAndRecreateDraft(*ph, opts)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Error deleting and recreating draft")

		return
	}
}

func getJsonBody(r *http.Request) []byte {
	body, _ := ioutil.ReadAll(r.Body)

	return body
}

func authorizeGithubHook(r *http.Request, jsonBody []byte) bool {
	secretKey := os.Getenv("GITHUB_HOOK_SECRET")

	if secretKey == "" {
		return false
	}

	hash := hmac.New(sha1.New, []byte(secretKey))
	if _, err := hash.Write(jsonBody); err != nil {
		return false
	}

	expectedHash := hex.EncodeToString(hash.Sum(nil))
	githubHash := strings.SplitN(r.Header.Get("X-Hub-Signature"), "=", 2)

	return expectedHash == githubHash[1]
}

func parseHookFromReq(r *http.Request, jsonBody []byte) (*PushHook, error) {
	var err error
	var ph PushHook

	err = json.Unmarshal([]byte(jsonBody), &ph)
	if err != nil {
		return nil, err
	}

	return &ph, nil
}

func deleteAndRecreateDraft(ph PushHook, opts *config.Options) error {
	baseBranchRef := "refs/heads/" + opts.BaseBranch
	if ph.Ref != baseBranchRef {
		return nil
	}

	gc, err := github.NewRepositoryClient(
		context.Background(),
		opts.AccToken,
		opts.Organization,
		opts.RepoName,
	)
	if err != nil {
		return err
	}

	version, err := release.NewVersion(gc, opts.BaseBranch).NextVersion(context.Background(), "", "", false, true)

	lastDraftRel, err := gc.LastDraftRelease(context.Background())
	if err == nil {
		err = gc.DeleteRelease(context.Background(), lastDraftRel)
		if err != nil {
			return err
		}
	}

	err = release.NewDraft(
		NewEmptyCLI(),
		gc,
		opts.BaseBranch,
		opts.ReleaseBodyPrefix,
		opts.ReleaseBranches,
		opts.ReleaseBodyBranches,
	).Create(context.Background(), version.Name, version.Name)

	return err
}