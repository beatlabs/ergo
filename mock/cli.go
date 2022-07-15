package mock

import (
	"sync"

	"github.com/beatlabs/ergo"
)

// CLI is a mock implementation.
type CLI struct {
	ConfirmationFn func() (bool, error)

	mu                sync.Mutex
	ConfirmationCalls int
}

// PrintTable is a mock implementation.
func (c *CLI) PrintTable(header []string, values [][]string) {}

// PrintColorizedLine is a mock implementation.
func (c *CLI) PrintColorizedLine(title, content string, level ergo.MessageLevel) {}

// PrintLine is a mock implementation.
func (c *CLI) PrintLine(content ...interface{}) {}

// Confirmation is a mock implementation.
func (c *CLI) Confirmation(actionText, cancellationMessage, successMessage string) (bool, error) {
	c.mu.Lock()
	c.ConfirmationCalls++
	c.mu.Unlock()
	if c.ConfirmationFn != nil {
		return c.ConfirmationFn()
	}
	return true, nil
}

// Input is a mock implementation.
func (c *CLI) Input() (string, error) {
	return "", nil
}
