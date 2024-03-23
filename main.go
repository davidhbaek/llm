package main

import (
	"os"

	"github.com/davidhbaek/llm/claude"
)

func main() {
	os.Exit(claude.CLI(os.Args[1:]))
}
