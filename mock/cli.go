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
	PrintTableCalls   []PrintTableVal
}

// PrintTableVal represents the values send to the PrintTable method.
type PrintTableVal struct {
	Header []string
	Values [][]string
}

// PrintTable is a mock implementation.
func (c *CLI) PrintTable(header []string, values [][]string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.PrintTableCalls = append(c.PrintTableCalls, PrintTableVal{Header: header, Values: values})
}

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
