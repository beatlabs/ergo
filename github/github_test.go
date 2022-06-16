package github

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/beatlabs/ergo"
	"github.com/google/go-github/v41/github"
)

func setup() (client *github.Client, mux *http.ServeMux, teardown func()) {
	const baseURLPath = "/api-v3"

	mux = http.NewServeMux()

	apiHandler := http.NewServeMux()
	apiHandler.Handle(baseURLPath+"/", http.StripPrefix(baseURLPath, mux))

	server := httptest.NewServer(apiHandler)
	client = github.NewClient(nil)
	clientURL, _ := url.Parse(server.URL + baseURLPath + "/")
	client.BaseURL = clientURL
	client.UploadURL = clientURL

	return client, mux, server.Close
}

func testBody(t *testing.T, r *http.Request, want string) {
	t.Helper()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("Error reading request body: %v", err)
	}
	if got := string(b); got != want {
		t.Errorf("request body\nhave: %q\nwant: %q", got, want)
	}
}

func testMethod(t *testing.T, r *http.Request, want string) {
	if want != r.Method {
		t.Errorf("Request method = %v, want %v", r.Method, want)
	}
}

func TestNewGithubClient(t *testing.T) {
	ctx := context.Background()
	client := NewGithubClient(ctx, "access_token")
	if client == nil {
		t.Fatalf("Client should not be nil")
	}
}

func TestNewRepositoryClientShouldReturnANewObject(t *testing.T) {
	client, _, teardown := setup()
	defer teardown()

	got := NewRepositoryClient("o", "r", client)

	if got == nil {
		t.Fatal("NewRepositoryClient should return a new RepositoryClient object.")
	}
}

func TestCreateDraftReleaseShouldCreateADraftRelease(t *testing.T) {
	ctx := context.Background()
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/releases", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{}`)
	})

	repClient := NewRepositoryClient("o", "r", client)

	err := repClient.CreateDraftRelease(ctx, "", "", "", "")
	if err != nil {
		t.Fatalf("CreateDraftRelease should not return the error: %v", err)
	}
}

func TestCompareBranchShouldCompareBranches(t *testing.T) {
	ctx := context.Background()
	client, mux, tearDown := setup()
	defer tearDown()

	mux.HandleFunc("/repos/o/r/compare/base...branch", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{ "commits": [{ "commit": { "message": "foo" } }] }`)
	})

	mux.HandleFunc("/repos/o/r/compare/branch...base", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{ "commits": [{ "commit": { "message": "foo" } }] }`)
	})

	repClient := NewRepositoryClient("o", "r", client)

	got, err := repClient.CompareBranch(ctx, "base", "branch")
	if err != nil {
		t.Fatalf("CompareBranch should not return the error: %v", err)
	}
	if got == nil {
		t.Fatalf("CompareBranch should return a StatusReport object.")
	}
	if got.BaseBranch != "base" {
		t.Fatalf("BaseBranch has wrong value")
	}
	if got.Branch != "branch" {
		t.Fatalf("Branch has wrong value")
	}
	if len(got.Ahead) != 1 {
		t.Fatalf("StatusReport.Ahead should has %d elements", len(got.Ahead))
	}
	if len(got.Behind) != 1 {
		t.Fatalf("StatusReport.Behind should has %d elements", len(got.Behind))
	}
}

func TestCompareBranchShouldReturnErrorForInvalidResposne(t *testing.T) {
	ctx := context.Background()
	client, mux, tearDown := setup()
	defer tearDown()

	mux.HandleFunc("/repos/o/r/compare/base...branch", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "invalid")
	})

	mux.HandleFunc("/repos/o/r/compare/branch...base", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{ "commits": [{ "commit": { "message": "foo" } }] }`)
	})

	repClient := NewRepositoryClient("o", "r", client)

	got, err := repClient.CompareBranch(ctx, "base", "branch")
	if err == nil {
		t.Fatal("CompareBranch should return the error for invalid response")
	}
	if got != nil {
		t.Error("CompareBranch should return nil on status code on error.")
	}
}

func TestCompareBranchShouldReturnErrorForInvalidResponseBranchBase(t *testing.T) {
	ctx := context.Background()
	client, mux, tearDown := setup()
	defer tearDown()

	mux.HandleFunc("/repos/o/r/compare/base...branch", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{ "commits": [{ "commit": { "message": "foo" } }] }`)
	})

	mux.HandleFunc("/repos/o/r/compare/branch...base", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "invalid")
	})

	repClient := NewRepositoryClient("o", "r", client)

	got, err := repClient.CompareBranch(ctx, "base", "branch")
	if err == nil {
		t.Fatal("CompareBranch should return the error for invalid response")
	}
	if got != nil {
		t.Error("CompareBranch should return nil on status code on error.")
	}
}

func TestLastReleaseShouldReturnTheLastRelease(t *testing.T) {
	ctx := context.Background()
	client, mux, tearDown := setup()
	defer tearDown()

	want := &ergo.Release{
		ID:         12,
		Body:       "release_body",
		TagName:    "release_tag_name",
		ReleaseURL: "url",
	}

	mux.HandleFunc("/repos/o/r/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{ "id": %d, "body": "%s", "tag_name": "%s", "html_url": "%s" }`, want.ID, want.Body, want.TagName, want.ReleaseURL)
	})

	repClient := NewRepositoryClient("o", "r", client)

	got, err := repClient.LastRelease(ctx)
	if err != nil {
		t.Errorf("LastRelease should not return the error: %v", err)
	}
	if *got != *want {
		t.Errorf("got = %v; want %v", *got, *want)
	}
}

func TestLastReleaseShouldReturnWrappedErrorForCodeStatusNotFound(t *testing.T) {
	ctx := context.Background()
	client, mux, tearDown := setup()
	defer tearDown()

	mux.HandleFunc("/repos/o/r/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	repClient := NewRepositoryClient("o", "r", client)

	_, err := repClient.LastRelease(ctx)
	var errorResponse *github.ErrorResponse
	if !errors.As(err, &errorResponse) {
		t.Fatalf("LastRelease(GitHub API responds with 404) returned wrapped error of type %T, want: %T,", err, errorResponse)
	}
	if got, want := errorResponse.Response.StatusCode, http.StatusNotFound; got != want {
		t.Errorf("LastRelease(GitHub API responds with 404) returned wrapped error of type %T with StatusCode %d, want: %d", errorResponse, got, want)
	}
}

func TestLastReleaseShouldNotReturnErrorForServerError(t *testing.T) {
	ctx := context.Background()
	client, mux, tearDown := setup()
	defer tearDown()

	mux.HandleFunc("/repos/o/r/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	repClient := NewRepositoryClient("o", "r", client)

	got, err := repClient.LastRelease(ctx)
	if err == nil {
		t.Error("Should return error for invalid internal server error")
	}
	if got != nil {
		t.Error("Release should be nil")
	}
}

func TestLastReleaseShouldReturnErrorForInvalidPayload(t *testing.T) {
	ctx := context.Background()
	client, mux, tearDown := setup()
	defer tearDown()

	mux.HandleFunc("/repos/o/r/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "invalid_payload")
	})

	repClient := NewRepositoryClient("o", "r", client)

	got, err := repClient.LastRelease(ctx)
	if err == nil {
		t.Error("Should return error for invalid payload")
	}
	if got != nil {
		t.Error("Release should be nil on error")
	}
}

func TestEditReleaseShouldEditTheRelease(t *testing.T) {
	ctx := context.Background()
	client, mux, tearDown := setup()
	defer tearDown()

	want := &ergo.Release{
		ID:         12,
		Body:       "release_body",
		TagName:    "tag_name",
		ReleaseURL: "release_url",
	}

	mux.HandleFunc("/repos/o/r/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{ "id": %d, "body": "%s", "tag_name": "%s", "html_url": "%s" }`, want.ID, want.Body, want.TagName, want.ReleaseURL)
	})
	mux.HandleFunc("/repos/o/r/releases/12", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{ "body": "%s" }`, want.Body)
	})

	repClient := NewRepositoryClient("o", "r", client)

	gotLatest, err := repClient.LastRelease(ctx)
	if err != nil {
		t.Fatalf("LastRelease should not return the error: %v", err)
	}
	if *gotLatest != *want {
		t.Fatalf("got = %v, want, %v", *gotLatest, *want)
	}

	gotRelease, err := repClient.EditRelease(ctx, want)
	if err != nil {
		t.Fatalf("EditRelease should not return the error: %v", err)
	}
	if gotRelease.Body != want.Body {
		t.Errorf("got = %s, want, %s", gotRelease.Body, want.Body)
	}
}

func TestEditReleaseShouldReturnErrorForNilInputRelease(t *testing.T) {
	ctx := context.Background()
	client, _, tearDown := setup()
	defer tearDown()

	repClient := NewRepositoryClient("o", "r", client)

	_, err := repClient.EditRelease(ctx, nil)
	if err == nil {
		t.Fatalf("EditRelease should return error when input release is nil")
	}
}

func TestEditReleaseShouldReturnErrorOnGitHubAPIError(t *testing.T) {
	ctx := context.Background()
	client, mux, tearDown := setup()
	defer tearDown()

	mux.HandleFunc("/repos/o/r/releases/12", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})

	repClient := NewRepositoryClient("o", "r", client)

	editRelease := &ergo.Release{
		ID:         12,
		Body:       "release_body",
		TagName:    "tag_name",
		ReleaseURL: "release_url",
	}
	_, err := repClient.EditRelease(ctx, editRelease)
	if err == nil {
		t.Errorf("EditRelease should return error when GitHub API sends code 503")
	}
}

func TestPublishReleaseSuccess(t *testing.T) {
	client, mux, tearDown := setup()
	defer tearDown()

	mux.HandleFunc("/repos/o/r/releases/12", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PATCH")
		testBody(t, r, `{"draft":false}`+"\n")
		fmt.Fprintf(w, `{}`)
	})

	repoClient := NewRepositoryClient("o", "r", client)

	ctx := context.Background()
	if err := repoClient.PublishRelease(ctx, 12); err != nil {
		t.Errorf("PublishRelease returned error: %v", err)
	}
}

func TestPublishReleaseError(t *testing.T) {
	client, mux, tearDown := setup()
	defer tearDown()

	mux.HandleFunc("/repos/o/r/releases/12", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PATCH")
		testBody(t, r, `{"draft":false}`+"\n")
		w.WriteHeader(http.StatusServiceUnavailable)
	})

	repoClient := NewRepositoryClient("o", "r", client)

	ctx := context.Background()
	if err := repoClient.PublishRelease(ctx, 12); err == nil {
		t.Error("PublishRelease expected to return error")
	}
}

func TestCreateTagShouldCreateANewTagForValidArgs(t *testing.T) {
	ctx := context.Background()
	client, mux, tearDown := setup()
	defer tearDown()

	want := &ergo.Tag{Name: "tag_name"}

	mux.HandleFunc("/repos/o/r/git/tags", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{ "tag": "%s" }`, want.Name)
	})
	mux.HandleFunc("/repos/o/r/git/refs", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{}`)
	})

	repClient := NewRepositoryClient("o", "r", client)

	got, err := repClient.CreateTag(ctx, "versionName", "SHA", "Message")
	if err != nil {
		t.Fatalf("Should not return the error: %v", err)
	}
	if *got != *want {
		t.Errorf("got = %v; want %v", got, want)
	}
}

func TestCreateTagShouldReturnErrorForInvalidRefResponse(t *testing.T) {
	ctx := context.Background()
	client, mux, tearDown := setup()
	defer tearDown()

	mux.HandleFunc("/repos/o/r/git/tags", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{ "tag": "tag_name" }`)
	})
	mux.HandleFunc("/repos/o/r/git/refs", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, "invalid_body")
	})

	repClient := NewRepositoryClient("o", "r", client)

	tag, err := repClient.CreateTag(ctx, "versionName", "SHA", "Message")
	if err == nil {
		t.Fatal("Should return error for invalid response")
	}
	if tag != nil {
		t.Error("Tag should be nil on error")
	}
}

func TestCreateTagShouldReturnErrorForInvalidTagResponse(t *testing.T) {
	ctx := context.Background()
	client, mux, tearDown := setup()
	defer tearDown()

	mux.HandleFunc("/repos/o/r/git/tags", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, "invalid_body")
	})

	repClient := NewRepositoryClient("o", "r", client)

	tag, err := repClient.CreateTag(ctx, "versionName", "SHA", "Message")
	if err == nil {
		t.Fatal("Should return error for invalid response")
	}
	if tag != nil {
		t.Error("Tag should be nil on error")
	}
}

func TestCreateTagShouldReturnErrorForInvalidRefResponseCode(t *testing.T) {
	ctx := context.Background()
	client, mux, tearDown := setup()
	defer tearDown()

	mux.HandleFunc("/repos/o/r/git/tags", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	repClient := NewRepositoryClient("o", "r", client)

	tag, err := repClient.CreateTag(ctx, "versionName", "SHA", "Message")
	if err != nil {
		t.Fatal("Should return error for invalid reference status code 404")
	}
	if tag != nil {
		t.Error("Tag should be nil on error")
	}
}

func TestCreateTagShouldNotReturnErrorWhenReferenceNotFound(t *testing.T) {
	ctx := context.Background()
	client, mux, tearDown := setup()
	defer tearDown()

	mux.HandleFunc("/repos/o/r/git/tags", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{ "tag": "tag_name" }`)
	})
	mux.HandleFunc("/repos/o/r/git/refs", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	repClient := NewRepositoryClient("o", "r", client)

	tag, err := repClient.CreateTag(ctx, "versionName", "SHA", "Message")
	if err != nil {
		t.Fatal("Should not return error for tag status code 404 ")
	}
	if tag != nil {
		t.Error("Tag should be nil on error")
	}
}

func TestDiffCommitsShouldReturnTheDiffsForValidInputs(t *testing.T) {
	ctx := context.Background()
	client, mux, tearDown := setup()
	defer tearDown()

	want := []*ergo.StatusReport{
		{
			Branch:     "base",
			BaseBranch: "branch",
			Ahead:      []*ergo.Commit{{Message: "foo"}},
			Behind:     []*ergo.Commit{{Message: "foo"}},
		},
	}

	mux.HandleFunc("/repos/o/r/compare/base...branch", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{ "commits": [{ "commit": { "message": "%s" } }] }`, want[0].Ahead[0].Message)
	})

	mux.HandleFunc("/repos/o/r/compare/branch...base", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{ "commits": [{ "commit": { "message": "%s" } }] }`, want[0].Behind[0].Message)
	})

	repClient := NewRepositoryClient("o", "r", client)

	got, err := repClient.DiffCommits(ctx, []string{"base"}, "branch")
	if err != nil {
		t.Fatalf("CompareBranch should not return the error: %v", err)
	}
	if got[0].Branch != want[0].Branch {
		t.Errorf("got = %v; want %v", got[0].Branch, want[0].Branch)
	}
	if got[0].BaseBranch != want[0].BaseBranch {
		t.Errorf("got = %v; want %v", got[0].BaseBranch, want[0].BaseBranch)
	}
	if *got[0].Ahead[0] != *want[0].Ahead[0] {
		t.Errorf("got = %v; want %v", *got[0].Ahead[0], *want[0].Ahead[0])
	}
	if *got[0].Behind[0] != *want[0].Behind[0] {
		t.Errorf("got = %v; want %v", *got[0].Behind[0], *want[0].Behind[0])
	}
}

func TestDiffCommitsShouldReturnErrorForInvalidPayload(t *testing.T) {
	ctx := context.Background()
	client, mux, tearDown := setup()
	defer tearDown()

	mux.HandleFunc("/repos/o/r/compare/base...branch", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "invalid_payload")
	})

	repClient := NewRepositoryClient("o", "r", client)

	diffCommits, err := repClient.DiffCommits(ctx, []string{"base"}, "branch")
	if err == nil {
		t.Fatal("DiffCommit should return error for invalid payload")
	}
	if diffCommits != nil {
		t.Error("DiffCommits should return nil on error")
	}
}

func TestUpdateBranchFromTagShouldUpdateBranchFromGivenTag(t *testing.T) {
	ctx := context.Background()
	client, mux, tearDown := setup()
	defer tearDown()

	mux.HandleFunc("/repos/o/r/git/ref/tags/test_tag", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"ref": "ref", "object": {"sha": "sha"}}`)
	})

	mux.HandleFunc("/repos/o/r/git/refs/heads/to_branch", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "{}")
	})

	repClient := NewRepositoryClient("o", "r", client)

	err := repClient.UpdateBranchFromTag(ctx, "test_tag", "to_branch", true)
	if err != nil {
		t.Fatalf("Should not return the error: %v", err)
	}
}

func TestUpdateBranchFromTagShouldReturnErrorForInvalidUpdateRefPayload(t *testing.T) {
	ctx := context.Background()
	client, mux, tearDown := setup()
	defer tearDown()

	mux.HandleFunc("/repos/o/r/git/ref/tags/test_tag", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"ref": "ref", "object": {"sha": "sha"}}`)
	})

	mux.HandleFunc("/repos/o/r/git/refs/heads/to_branch", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "invalid_payload")
	})

	repClient := NewRepositoryClient("o", "r", client)

	err := repClient.UpdateBranchFromTag(ctx, "test_tag", "to_branch", true)
	if err == nil {
		t.Fatal("UpdateBranchFromTag should return error for invalid response")
	}
}

func TestUpdateBranchFromTagShouldReturnErrorForInvalidGetRefPayload(t *testing.T) {
	ctx := context.Background()
	client, mux, tearDown := setup()
	defer tearDown()

	mux.HandleFunc("/repos/o/r/git/ref/tags/test_tag", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "invalid_payload")
	})

	mux.HandleFunc("/repos/o/r/git/refs/heads/to_branch", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "{}")
	})

	repClient := NewRepositoryClient("o", "r", client)

	err := repClient.UpdateBranchFromTag(ctx, "test_tag", "to_branch", true)
	if err == nil {
		t.Fatal("UpdateBranchFromTag should return error for invalid response")
	}
}

func TestGetRefFromTagShouldGetAReferenceFromTag(t *testing.T) {
	ctx := context.Background()
	client, mux, tearDown := setup()
	defer tearDown()

	want := &ergo.Reference{SHA: "sha", Ref: "ref"}

	mux.HandleFunc("/repos/o/r/git/ref/tags/test_tag", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"ref": "%s", "object": {"sha": "%s"}}`, want.Ref, want.SHA)
	})

	repClient := NewRepositoryClient("o", "r", client)

	got, err := repClient.GetRefFromTag(ctx, "test_tag")
	if err != nil {
		t.Fatalf("Should not return the error: %v", err)
	}
	if *got != *want {
		t.Errorf("got = %v; want %v", *got, *want)
	}
}

func TestGetRefFromTagShouldReturnErrorForInvalidBody(t *testing.T) {
	ctx := context.Background()
	client, mux, tearDown := setup()
	defer tearDown()

	mux.HandleFunc("/repos/o/r/git/ref/tags/test_tag", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "invalid_body")
	})

	repClient := NewRepositoryClient("o", "r", client)

	got, err := repClient.GetRefFromTag(ctx, "test_tag")
	if err == nil {
		t.Fatal("Should return error for invalid body")
	}
	if got != nil {
		t.Error("Should return nil Reference on error")
	}
}

func TestGetRefFromTagShouldReturnNilForStatusNotFound(t *testing.T) {
	ctx := context.Background()
	client, mux, tearDown := setup()
	defer tearDown()

	mux.HandleFunc("/repos/o/r/git/ref/tags/test_tag", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	repClient := NewRepositoryClient("o", "r", client)

	got, err := repClient.GetRefFromTag(ctx, "test_tag")
	if err != nil {
		t.Fatalf("Should not return error for status not found, error: %v", err)
	}
	if got != nil {
		t.Error("Should return nil on status not found")
	}
}

func TestGetRefShouldReturnTheReference(t *testing.T) {
	ctx := context.Background()
	client, mux, tearDown := setup()
	defer tearDown()

	want := &ergo.Reference{SHA: "sha", Ref: "ref"}

	mux.HandleFunc("/repos/o/r/git/ref/heads/test", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"ref": "%s", "object": {"sha": "%s"}}`, want.Ref, want.SHA)
	})

	repClient := NewRepositoryClient("o", "r", client)

	got, err := repClient.GetRef(ctx, "test")
	if err != nil {
		t.Fatalf("Should not return the error: %v", err)
	}

	if *got != *want {
		t.Errorf("got = %v; want %v", *got, *want)
	}
}

func TestGetRefShouldReturnErrorForInvalidBody(t *testing.T) {
	ctx := context.Background()
	client, mux, tearDown := setup()
	defer tearDown()

	mux.HandleFunc("/repos/o/r/git/ref/heads/test", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "invalid_body")
	})

	repClient := NewRepositoryClient("o", "r", client)

	got, err := repClient.GetRef(ctx, "test")
	if err == nil {
		t.Fatal("Should return error for invalid body")
	}
	if got != nil {
		t.Error("Ref should be nil on error")
	}
}

func TestGetRepoNameShouldReturnTheRepoName(t *testing.T) {
	client, _, tearDown := setup()
	defer tearDown()

	repClient := NewRepositoryClient("o", "r", client)

	want := "o/r"
	got := repClient.GetRepoName()
	if got != want {
		t.Errorf("got = %s, want = %s", got, want)
	}
}
