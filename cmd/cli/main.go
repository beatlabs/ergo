package main

import (
	"github.com/beatlabs/ergo/commands"
)

var version = "develop"

func main() {
	commands.Execute(version)
}
