package release_test

import (
	"context"
	"errors"
	"testing"

	"github.com/thebeatapp/ergo"
	"github.com/thebeatapp/ergo/mock"
	"github.com/thebeatapp/ergo/release"
)

var (
	ctx context.Context
)

func TestNewTagShouldNotReturnNilObject(t *testing.T) {
	var host ergo.Host
	if release.NewTag(host) == nil {
		t.Error("expected Tag object to not be nil.")
	}
}

func TestCreateShouldCreateTag(t *testing.T) {
	host := &mock.RepositoryClient{}
	versionName := "1.0.0"

	want := ergo.Tag{Name: versionName}

	host.MockCreateTagFn = func() (*ergo.Tag, error) {
		return &ergo.Tag{Name: versionName}, nil
	}

	got, err := release.NewTag(host).Create(ctx, &ergo.Version{Name: versionName, SHA: "sha"})
	if err != nil {
		t.Fatalf("error creating tag: %v", err)
	}

	if want != *got {
		t.Errorf("expected created tag object to be equal to %v", want)
	}
}

func TestCreateShouldReturnErrorForDifferentVersionNames(t *testing.T) {
	host := &mock.RepositoryClient{}
	want := ergo.Tag{Name: "1.0.0"}

	host.MockCreateTagFn = func() (*ergo.Tag, error) {
		return &ergo.Tag{Name: "2.0.0"}, nil
	}

	got, err := release.NewTag(host).Create(ctx, &ergo.Version{Name: "2.0.0", SHA: "sha"})
	if err != nil {
		t.Fatalf("error creating tag: %v", err)
	}

	if want == *got {
		t.Errorf("expected created tag object to be equal to %v", want)
	}
}

func TestCreateShouldReturnError(t *testing.T) {
	host := &mock.RepositoryClient{}
	host.MockCreateTagFn = func() (*ergo.Tag, error) {
		return nil, errors.New("")
	}

	version := &ergo.Version{Name: "s", SHA: "sha"}
	if _, err := release.NewTag(host).Create(ctx, version); err == nil {
		t.Errorf("expected Create to return error")
	}
}

func TestExistsTagNameShouldReturnTrue(t *testing.T) {
	host := &mock.RepositoryClient{}

	host.MockGetRefFromTagFn = func() (*ergo.Reference, error) {
		return &ergo.Reference{SHA: "sha", Ref: "ref"}, nil
	}

	got, err := release.NewTag(host).ExistsTagName(ctx, "name")
	if err != nil {
		t.Fatalf("error checking for tag name: %v", err)
	}

	if !got {
		t.Errorf("expected ExistsTagName to be %v", true)
	}
}

func TestExistsTagNameShouldReturnFalse(t *testing.T) {
	host := &mock.RepositoryClient{}

	got, err := release.NewTag(host).ExistsTagName(ctx, "name")
	if err != nil {
		t.Fatalf("error checking for tag name: %v", err)
	}

	if got {
		t.Errorf("expected ExistsTagName to be %v", false)
	}
}

func TestExistsTagNameShouldReturnError(t *testing.T) {
	host := &mock.RepositoryClient{}
	host.MockGetRefFromTagFn = func() (*ergo.Reference, error) {
		return nil, errors.New("")
	}

	_, err := release.NewTag(host).ExistsTagName(ctx, "name")
	if err == nil {
		t.Fatalf("expected ExistsTagName to return error")
	}
}
