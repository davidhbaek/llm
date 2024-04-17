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

### Provide a PDF as context

```
$./llm -d <path/to/pdf> -p "summarize this document" -m gpt4
```

### Start a chat session

```
$./llm -c -m gpt4
```
### Flags
- `-p, --prompt`: user prompt
- `-s, --system`: system prompt
- `-i, --image`: filepath or URL of image
- `-d, --document`: filepath of document (PDF)
- `-m, --model`: name of LLM to use [gpt4, haiku, sonnet, opus]
- `-c, --chat`: start an interactive chat session


