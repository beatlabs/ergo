package release

import (
	"errors"
	"testing"

	"github.com/beatlabs/ergo"
	"github.com/beatlabs/ergo/mock"
)

func TestNewVersionShouldNotReturnNilObject(t *testing.T) {
	var host ergo.Host
	if NewVersion(host, "baseBranch") == nil {
		t.Error("expected Tag object to not be nil.")
	}
}

func TestNextVersionShouldReturnTheNextVersionWithDefaultParameters(t *testing.T) {
	host := &mock.RepositoryClient{}
	host.MockGetRefFn = func() (*ergo.Reference, error) {
		return &ergo.Reference{SHA: "sha", Ref: "ref"}, nil
	}
	host.MockLastReleaseFn = func() (*ergo.Release, error) {
		return &ergo.Release{TagName: "1.0.0"}, nil
	}

	want := ergo.Version{
		SHA:  "sha",
		Name: "1.0.1",
	}

	got, err := NewVersion(host, "baseBranch").
		NextVersion(ctx, "", "", false, false)
	if err != nil {
		t.Fatal(err)
	}

	if want != *got {
		t.Errorf("expected next version to be equal to %v", want)
	}
}

func TestNextVersionShouldReturnTheNextVersionWithSuffixSameSHA(t *testing.T) {
	host := &mock.RepositoryClient{}
	host.MockGetRefFn = func() (*ergo.Reference, error) {
		return &ergo.Reference{SHA: "sha", Ref: "ref"}, nil
	}
	host.MockLastReleaseFn = func() (*ergo.Release, error) {
		return &ergo.Release{TagName: "1.0.0"}, nil
	}
	host.MockGetRefFromTagFn = func() (*ergo.Reference, error) {
		return &ergo.Reference{SHA: "sha", Ref: "ref"}, nil
	}

	want := ergo.Version{
		SHA:  "sha",
		Name: "1.0.0-mx",
	}

	got, err := NewVersion(host, "baseBranch").
		NextVersion(ctx, "", "mx", false, false)
	if err != nil {
		t.Fatal(err)
	}

	if want != *got {
		t.Errorf("expected next version to be equal to %v instead of %v", want, *got)
	}
}

func TestNextVersionShouldReturnTheNextVersionWithSuffixDifferentSHA(t *testing.T) {
	host := &mock.RepositoryClient{}
	host.MockGetRefFn = func() (*ergo.Reference, error) {
		return &ergo.Reference{SHA: "sha1", Ref: "ref"}, nil
	}
	host.MockLastReleaseFn = func() (*ergo.Release, error) {
		return &ergo.Release{TagName: "1.0.0"}, nil
	}
	host.MockGetRefFromTagFn = func() (*ergo.Reference, error) {
		return &ergo.Reference{SHA: "sha2", Ref: "ref"}, nil
	}

	want := ergo.Version{
		SHA:  "sha1",
		Name: "1.0.1-mx",
	}

	got, err := NewVersion(host, "baseBranch").
		NextVersion(ctx, "", "mx", false, false)
	if err != nil {
		t.Fatal(err)
	}

	if want != *got {
		t.Errorf("expected next version to be equal to %v instead of %v", want, *got)
	}
}

func TestNextVersionShouldReturnTheNextVersionWithInputVersion(t *testing.T) {
	host := &mock.RepositoryClient{}
	host.MockGetRefFn = func() (*ergo.Reference, error) {
		return &ergo.Reference{SHA: "sha", Ref: "ref"}, nil
	}

	want := ergo.Version{
		SHA:  "sha",
		Name: "v13.0.5.1-custom",
	}

	got, err := NewVersion(host, "baseBranch").
		NextVersion(ctx, "v13.0.5.1-custom", "", false, false)
	if err != nil {
		t.Fatal(err)
	}

	if want != *got {
		t.Errorf("expected next version to be equal to %v instead of %v", want, *got)
	}
}

func TestNextVersionShouldReturnTheNextVersionWithInputVersionAndSuffix(t *testing.T) {
	host := &mock.RepositoryClient{}
	host.MockGetRefFn = func() (*ergo.Reference, error) {
		return &ergo.Reference{SHA: "sha", Ref: "ref"}, nil
	}

	want := ergo.Version{
		SHA:  "sha",
		Name: "v13.0.5.1-custom",
	}

	got, err := NewVersion(host, "baseBranch").
		NextVersion(ctx, "v13.0.5.1", "custom", false, false)
	if err != nil {
		t.Fatal(err)
	}

	if want != *got {
		t.Errorf("expected next version to be equal to %v instead of %v", want, *got)
	}
}

func TestNextVersionShouldReturnTheNextVersionWithCustomLastReleaseVersion(t *testing.T) {
	host := &mock.RepositoryClient{}
	host.MockGetRefFn = func() (*ergo.Reference, error) {
		return &ergo.Reference{SHA: "sha", Ref: "ref"}, nil
	}
	host.MockLastReleaseFn = func() (*ergo.Release, error) {
		return &ergo.Release{TagName: "v13.0.5.1-custom"}, nil
	}

	want := ergo.Version{
		SHA:  "sha",
		Name: "0.0.1",
	}

	got, err := NewVersion(host, "baseBranch").
		NextVersion(ctx, "", "", false, false)
	if err != nil {
		t.Fatal(err)
	}

	if want != *got {
		t.Errorf("expected next version to be equal to %v instead of %v", want, *got)
	}
}

func TestNextVersionShouldReturnTheNextVersionWithMinorFlag(t *testing.T) {
	host := &mock.RepositoryClient{}
	host.MockGetRefFn = func() (*ergo.Reference, error) {
		return &ergo.Reference{SHA: "sha", Ref: "ref"}, nil
	}
	host.MockLastReleaseFn = func() (*ergo.Release, error) {
		return &ergo.Release{TagName: "1.0.0"}, nil
	}

	want := ergo.Version{
		SHA:  "sha",
		Name: "1.1.0",
	}

	got, err := NewVersion(host, "baseBranch").
		NextVersion(ctx, "", "", false, true)
	if err != nil {
		t.Fatal(err)
	}

	if want != *got {
		t.Errorf("expected next version to be equal to %v instead of %v", want, *got)
	}
}

func TestNextVersionShouldReturnTheNextVersionWithMajorFlag(t *testing.T) {
	host := &mock.RepositoryClient{}
	host.MockGetRefFn = func() (*ergo.Reference, error) {
		return &ergo.Reference{SHA: "sha", Ref: "ref"}, nil
	}
	host.MockLastReleaseFn = func() (*ergo.Release, error) {
		return &ergo.Release{TagName: "1.0.0"}, nil
	}

	want := ergo.Version{
		SHA:  "sha",
		Name: "2.0.0",
	}

	got, err := NewVersion(host, "baseBranch").
		NextVersion(ctx, "", "", true, false)
	if err != nil {
		t.Fatal(err)
	}

	if want != *got {
		t.Errorf("expected next version to be equal to %v instead of %v", want, *got)
	}
}

func TestNextVersionShouldReturnTheNextVersionWithSuffixAndMajor(t *testing.T) {
	host := &mock.RepositoryClient{}
	host.MockGetRefFn = func() (*ergo.Reference, error) {
		return &ergo.Reference{SHA: "sha1", Ref: "ref"}, nil
	}
	host.MockLastReleaseFn = func() (*ergo.Release, error) {
		return &ergo.Release{TagName: "1.0.0"}, nil
	}
	host.MockGetRefFromTagFn = func() (*ergo.Reference, error) {
		return &ergo.Reference{SHA: "sha2", Ref: "ref"}, nil
	}

	want := ergo.Version{
		SHA:  "sha1",
		Name: "2.0.0-mx",
	}

	got, err := NewVersion(host, "baseBranch").
		NextVersion(ctx, "", "mx", true, false)
	if err != nil {
		t.Fatal(err)
	}

	if want != *got {
		t.Errorf("expected next version to be equal to %v instead of %v", want, *got)
	}
}

func TestNextVersionShouldReturnTheNextVersionWithMajorAndMinor(t *testing.T) {
	host := &mock.RepositoryClient{}
	host.MockGetRefFn = func() (*ergo.Reference, error) {
		return &ergo.Reference{SHA: "sha", Ref: "ref"}, nil
	}
	host.MockLastReleaseFn = func() (*ergo.Release, error) {
		return &ergo.Release{TagName: "1.0.0"}, nil
	}

	want := ergo.Version{
		SHA:  "sha",
		Name: "2.0.0",
	}

	got, err := NewVersion(host, "baseBranch").
		NextVersion(ctx, "", "", true, true)
	if err != nil {
		t.Fatal(err)
	}

	if want != *got {
		t.Errorf("expected next version to be equal to %v instead of %v", want, *got)
	}
}

func TestNextVersionShouldReturnDefaultVersionWhenNoReleases(t *testing.T) {
	host := &mock.RepositoryClient{}
	host.MockGetRefFn = func() (*ergo.Reference, error) {
		return &ergo.Reference{SHA: "sha", Ref: "ref"}, nil
	}
	host.MockLastReleaseFn = func() (*ergo.Release, error) {
		return nil, nil
	}

	want := ergo.Version{
		SHA:  "sha",
		Name: "0.0.1",
	}

	got, err := NewVersion(host, "baseBranch").
		NextVersion(ctx, "", "", false, false)
	if err != nil {
		t.Fatal(err)
	}

	if want != *got {
		t.Errorf("expected next version to be equal to %v instead of %v", want, *got)
	}
}

func TestNextVersionShouldReturnErrorOnGetRef(t *testing.T) {
	host := &mock.RepositoryClient{}
	host.MockGetRefFn = func() (*ergo.Reference, error) {
		return nil, errors.New("")
	}

	_, err := NewVersion(host, "baseBranch").
		NextVersion(ctx, "", "", false, false)
	if err == nil {
		t.Error("expected NextVersion to return error")
	}
}

func TestNextVersionShouldReturnErrorOnLastRelease(t *testing.T) {
	host := &mock.RepositoryClient{}
	host.MockGetRefFn = func() (*ergo.Reference, error) {
		return &ergo.Reference{}, nil
	}
	host.MockLastReleaseFn = func() (*ergo.Release, error) {
		return nil, errors.New("")
	}

	_, err := NewVersion(host, "baseBranch").
		NextVersion(ctx, "", "", false, false)
	if err == nil {
		t.Error("expected NextVersion to return error")
	}
}

func TestNextVersionShouldReturnDefaultVersionOnWrongFormat(t *testing.T) {
	host := &mock.RepositoryClient{}
	host.MockGetRefFn = func() (*ergo.Reference, error) {
		return &ergo.Reference{SHA: "sha"}, nil
	}
	host.MockLastReleaseFn = func() (*ergo.Release, error) {
		return &ergo.Release{TagName: "wrong"}, nil
	}

	want := ergo.Version{
		SHA:  "sha",
		Name: "0.0.1",
	}

	got, err := NewVersion(host, "baseBranch").
		NextVersion(ctx, "", "", false, false)
	if err != nil {
		t.Fatal(err)
	}

	if want != *got {
		t.Errorf("expected next version to be equal to: %v instead of %v", want, *got)
	}
}

func TestNextVersionShouldReturnErrorOnLastReleaseWithSuffix(t *testing.T) {
	host := &mock.RepositoryClient{}
	host.MockGetRefFn = func() (*ergo.Reference, error) {
		return &ergo.Reference{SHA: "sha", Ref: "ref"}, nil
	}
	host.MockLastReleaseFn = func() (*ergo.Release, error) {
		return &ergo.Release{TagName: "1.0.0"}, nil
	}
	host.MockGetRefFromTagFn = func() (*ergo.Reference, error) {
		return nil, errors.New("")
	}

	_, err := NewVersion(host, "baseBranch").
		NextVersion(ctx, "", "-mx", false, false)
	if err == nil {
		t.Error("expected NextVersion to return error")
	}
}
