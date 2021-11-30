package ergo

import "context"

// MessageLevel defines the level of output message.
type MessageLevel string

// Host interface describes the host's actions.
type Host interface {
	CreateDraftRelease(ctx context.Context, name, tagName, releaseBody, targetBranch string) error
	LastRelease(ctx context.Context) (*Release, error)
	EditRelease(ctx context.Context, release *Release) (*Release, error)
	CompareBranch(ctx context.Context, baseBranch, branch string) (*StatusReport, error)
	DiffCommits(ctx context.Context, releaseBranches []string, baseBranch string) ([]*StatusReport, error)
	CreateTag(ctx context.Context, versionName, sha, m string) (*Tag, error)
	UpdateBranchFromTag(ctx context.Context, tag, toBranch string, force bool) error
	GetRef(ctx context.Context, branch string) (*Reference, error)
	GetRefFromTag(ctx context.Context, tag string) (*Reference, error)
	GetRepoName() string
}

// CLI describes the command line interface actions.
type CLI interface {
	PrintTable(header []string, values [][]string)
	PrintColorizedLine(title, content string, level MessageLevel)
	PrintLine(content ...interface{})
	Confirmation(actionText, cancellationMessage, successMessage string) (bool, error)
	Input() (string, error)
}

// Deploy describes the deploy process.
type Deploy interface {
	Do(ctx context.Context, releaseIntervalInput, releaseOffsetInput string, allowForcePush bool) error
}

// Draft describes the draft process.
type Draft interface {
	Create(ctx context.Context, releaseName, tagName string) error
}

// Release struct contains all the fields which describe the release entity.
type Release struct {
	ID         int64
	Body       string
	TagName    string
	ReleaseURL string
}

// StatusReport struct is responsible to keep the information about current status.
type StatusReport struct {
	Branch     string
	BaseBranch string
	Ahead      []*Commit
	Behind     []*Commit
}

// Commit describes the commit entity.
type Commit struct {
	Message string
}

// Tag describes the tag entity.
type Tag struct {
	Name string
}

// Reference describes the reference entity.
type Reference struct {
	SHA string
	Ref string
}

// Version describe the version entity.
type Version struct {
	Name string
	SHA  string
}
