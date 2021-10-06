package release

import (
	"context"
	"fmt"
	"strings"

	"github.com/beatlabs/ergo"
	"github.com/beatlabs/ergo/cli"
	"github.com/pkg/errors"
)

// Draft is responsible for creating the draft release.
type Draft struct {
	c                   ergo.CLI
	host                ergo.Host
	baseBranch          string
	releaseBodyPrefix   string
	releaseBranches     []string
	releaseBodyBranches map[string]string
}

// NewDraft initialize and return a new Draft object.
func NewDraft(
	c ergo.CLI,
	host ergo.Host,
	baseBranch, releaseBodyPrefix string,
	releaseBranches []string,
	releaseBodyBranches map[string]string,
) *Draft {
	return &Draft{
		c:                   c,
		host:                host,
		baseBranch:          baseBranch,
		releaseBodyPrefix:   releaseBodyPrefix,
		releaseBranches:     releaseBranches,
		releaseBodyBranches: releaseBodyBranches,
	}
}

// Create is responsible to create a new draft release.
func (d *Draft) Create(ctx context.Context, releaseName, tagName string) error {
	diff, err := d.host.DiffCommits(ctx, d.releaseBranches, d.baseBranch)
	if err != nil {
		return err
	}

	releaseBody := d.releaseBody(diff, d.releaseBodyPrefix, d.releaseBodyBranches)

	d.c.PrintColorizedLine("REPO: ", d.host.GetRepoName(), cli.WarningType)
	d.c.PrintLine(releaseBody)

	confirm, err := d.c.Confirmation(
		"Draft the release",
		"No draft",
		"The draft release is ready",
	)

	if err != nil {
		return errors.Wrap(err, "confirmation dialog error")
	}

	if !confirm {
		return nil
	}

	return d.host.CreateDraftRelease(ctx, releaseName, tagName, releaseBody)
}

// releaseBody output needed for github release body.
func (d *Draft) releaseBody(commitDiffBranches []*ergo.StatusReport, releaseBodyPrefix string, branchMap map[string]string) string {
	var formattedCommits []string
	var formattedBranches []string
	var header, body string

	firstLinePrefix := "- "
	nextLinePrefix := "  "
	lineSeparator := "\r\n"

	for _, diffBranch := range commitDiffBranches {
		branchText, ok := branchMap[diffBranch.Branch]
		if !ok {
			branchText = diffBranch.Branch
		}
		formattedBranches = append(formattedBranches,
			fmt.Sprintf("%s ![](https://img.shields.io/badge/released-No-red.svg)", branchText))
	}

	if len(commitDiffBranches) >= 1 {
		for _, commit := range commitDiffBranches[0].Behind {
			formattedCommits = append(formattedCommits, d.formatMessage(commit, firstLinePrefix, nextLinePrefix, lineSeparator))
			body = fmt.Sprintf("%s%s%s",
				body,
				d.formatMessage(commit, firstLinePrefix, nextLinePrefix, lineSeparator),
				lineSeparator)
		}
	}

	header = strings.Join(formattedBranches, " ")
	body = strings.Join(formattedCommits, lineSeparator)
	parts := []string{header, releaseBodyPrefix, body}

	return strings.Join(parts, strings.Repeat(lineSeparator, 2))
}

// formatMessage formats the commit's message.
func (d *Draft) formatMessage(c *ergo.Commit, firstLinePrefix, nextLinesPrefix, lineSeparator string) string {
	var outputStrings []string
	maxLines := 6

	lines := strings.Split(c.Message, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var prefix string

		if len(outputStrings) == 0 {
			prefix = firstLinePrefix
		} else {
			prefix = nextLinesPrefix
		}

		outputStrings = append(outputStrings, fmt.Sprintf("%s%s", prefix, strings.TrimSpace(line)))

		if len(outputStrings) >= maxLines {
			break
		}
	}
	return strings.Join(outputStrings, lineSeparator)
}
