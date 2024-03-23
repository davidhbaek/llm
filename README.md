# llm

A Go CLI tool to interact with LLMs

## Usage

### Flags
- `-p, --prompt`: user prompt
- `-s, --system`: system prompt
- `i, --image`: filepath or URL of image
- `-d, --document`: filepath of document (PDF)
- `-m, --model`: name of LLM to use (defaults to Claude's Haiku)


### Examples

#### Analyze a PDF

```
$ ./llm -m opus -d './prompts/docs/scaling-chatgpt.pdf' -p 'Analyze this document and provide a summary and key takeaways'
```

#### Analyze an image and provide a role via a system prompt

```
$ ./llm -i './prompts/images/ski-lodge.png'  -s 'You are a world renowned poet' -p 'Describe this image'
```
