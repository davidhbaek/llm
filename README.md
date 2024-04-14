# llm

A Go CLI tool to interact with LLMs

## Installation

### Pre-requisites

- [Go](https://go.dev/dl/)



## Usage

### Add your API keys

```
$ export ANTHROPIC_API_KEY=<your anthropic key>
$ export OPENAI_API_KEY=<your openai key>
```

### Build the executable

```
$ go build -o llm cmd/llm/main.go
```

### Send a prompt 

```
$ ./llm -m haiku -p hello
```

### Flags
- `-p, --prompt`: user prompt
- `-s, --system`: system prompt
- `-i, --image`: filepath or URL of image
- `-d, --document`: filepath of document (PDF)
- `-m, --model`: name of LLM to use (defaults to Claude's Haiku)


