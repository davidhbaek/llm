# llm

A Go CLI tool to interact with LLMs

## Installation

### Pre-requisites

- [Go](https://go.dev/dl/)
- API keys for your LLM provider

## Usage

### Populate your API keys

Add your API keys for the LLM provider you want to use

Models currently supported are from [OpenAI](https://openai.com/) and [Anthropic](https://www.anthropic.com/)

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
$ ./llm -m gpt4 -p hello
```

### Provide a PDF as context

```
$./llm -m gpt4 -d <path/to/pdf> -p "summarize this document"
```

### Start a chat session

```
$./llm -m gpt4 -c
```

### Flags
- `-p, --prompt`: user prompt
- `-s, --system`: system prompt
- `-i, --image`: filepath or URL of image
- `-d, --document`: filepath of document (PDF)
- `-m, --model`: name of LLM to use [gpt4, haiku, sonnet, opus]
- `-c, --chat`: start an interactive chat session


