package main

import "github.com/taxibeat/ergo"

func NewEmptyCLI() *EmptyCLI {
	return &EmptyCLI{}
}

type EmptyCLI struct{}

func (EmptyCLI) PrintTable(header []string, values [][]string) {
}

func (EmptyCLI) PrintColorizedLine(title, content string, level ergo.MessageLevel) {
}

func (EmptyCLI) PrintLine(content ...interface{}) {
}

func (EmptyCLI) Confirmation(actionText, cancellationMessage, successMessage string) (bool, error) {
	return true, nil
}

func (EmptyCLI) Input() (string, error) {
	return "", nil
}
