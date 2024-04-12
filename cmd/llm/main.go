package main

import (
	"os"

	"github.com/davidhbaek/llm/internal/llm"
)

func main() {
	llm.CLI(os.Args[1:])
}
