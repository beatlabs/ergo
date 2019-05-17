package cli

import (
	"bufio"
	"fmt"
	"os"

	"github.com/beatlabs/ergo"
	"github.com/fatih/color"
	"github.com/rodaine/table"
)

var (
	// InfoType defines the info type color.
	InfoType ergo.MessageLevel = "blue"
	// WarningType defines the warning color.
	WarningType ergo.MessageLevel = "yellow"
	// ErrorType defines the error color.
	ErrorType ergo.MessageLevel = "red"
	// SuccessType defines the success color.
	SuccessType ergo.MessageLevel = "green"

	blue   = color.New(color.FgCyan)
	yellow = color.New(color.FgYellow)
	red    = color.New(color.FgRed)
	white  = color.New(color.FgWhite)
	green  = color.New(color.FgGreen)

	confirmationText     = "[y/N]: "
	confirmationResponse = []string{"y", "Y"}
)

// CLI struct prints the message to console.
type CLI struct {
}

// NewCLI initialize and return an new CLI object.
func NewCLI() *CLI {
	return &CLI{}
}

// PrintTable is responsible to print a table view to terminal.
func (CLI) PrintTable(header []string, values [][]string) {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := yellow.SprintfFunc()

	convertedHeaders := convertToArrayOfInterface(header)
	tbl := table.New(convertedHeaders...)

	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	// CLI the body
	for _, val := range values {
		convertedValue := convertToArrayOfInterface(val)
		tbl.AddRow(convertedValue...)
	}

	tbl.Print()
}

// PrintColorizedLine print a colorized line.
func (CLI) PrintColorizedLine(title, content string, level ergo.MessageLevel) {
	var err error
	var clr *color.Color

	switch level {
	case ErrorType:
		clr = red
	case WarningType:
		clr = yellow
	case InfoType:
		clr = blue
	case SuccessType:
		clr = green
	default:
		clr = white
	}

	_, err = blue.Print(title)
	if err != nil {
		fmt.Print(title)
	}
	_, err = clr.Println(content)
	if err != nil {
		fmt.Println(content)
	}
}

// PrintLine print a line.
func (CLI) PrintLine(content ...interface{}) {
	_, err := white.Println(content...)
	if err != nil {
		fmt.Println(content...)
	}
}

// Confirmation will return true if user press ok otherwise will return false.
func (c CLI) Confirmation(actionText, cancellationMessage, successMessage string) (bool, error) {
	_, err := yellow.Print(actionText, "? ", confirmationText)
	if err != nil {
		return false, err
	}

	input, err := c.Input()
	if err != nil {
		return false, err
	}

	if !inSlice(input, confirmationResponse) {
		if cancellationMessage != "" {
			_, err = red.Println(cancellationMessage)
		}
		return false, err
	}
	if successMessage != "" {
		_, err = green.Println(successMessage)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

// Input reads a line from standard input and returns it.
func (CLI) Input() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	input, _, err := reader.ReadLine()

	if err != nil {
		return "", err
	}

	return string(input), nil
}

// convertToArrayOfInterface convert the given values to array of interfaces.
func convertToArrayOfInterface(values []string) []interface{} {
	var results []interface{}
	for _, value := range values {
		results = append(results, value)
	}
	return results
}

// inSlice check if the string exists in slice.
func inSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
