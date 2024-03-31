# llm

A Go CLI tool to interact with LLMs

## Usage

### Add your API keys

Create a `.env` file at the root of the repository and add your `ANTHROPIC_API_KEY`

### Flags
- `-p, --prompt`: user prompt
- `-s, --system`: system prompt
- `-i, --image`: filepath or URL of image
- `-d, --document`: filepath of document (PDF)
- `-m, --model`: name of LLM to use (defaults to Claude's Haiku)


### Examples

#### Analyze a PDF

```
$ ./llm -m opus -d '<path/to/pdf>' -p 'Analyze this document and provide a summary and key takeaways'
```

#### Analyze an image and provide a role via a system prompt

```
$ ./llm -i '<path/to/local_or_hosted_img>'  -s 'You are a world renowned poet' -p 'Describe this image'
```
