package main

import (
	"os"

	"github.com/davidhbaek/llm/internal/claude"
)

func main() {
	os.Exit(claude.CLI(os.Args[1:]))
}
