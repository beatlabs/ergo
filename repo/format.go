package repo

import (
	"fmt"
	"strings"

	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

// FormatMessage formats the commit's message
func FormatMessage(c *object.Commit, firstLinePrefix string, nextLinesPrefix string, lineSeparator string) string {
	outputStrings := []string{}
	maxLines := 6

	lines := strings.Split(c.Message, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		prefix := ""

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
