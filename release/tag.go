package release

import (
	"context"

	"github.com/thebeatapp/ergo"
)

// Tag describes the actions of tag entity.
type Tag struct {
	host ergo.Host
}

// NewTag initialize and return a new tag object.
func NewTag(host ergo.Host) *Tag {
	return &Tag{host: host}
}

// Create a new tag.
func (t Tag) Create(ctx context.Context, version *ergo.Version) (*ergo.Tag, error) {
	tag, err := t.host.CreateTag(ctx, version.Name, version.SHA, "")
	if err != nil {
		return nil, err
	}

	return &ergo.Tag{Name: tag.Name}, nil
}

// ExistsTagName checks if a tag exists by its name
func (t Tag) ExistsTagName(ctx context.Context, tagName string) (bool, error) {
	ref, err := t.host.GetRefFromTag(ctx, tagName)
	if err != nil {
		return false, err
	}

	if ref != nil {
		return true, nil
	}

	return false, nil
}
