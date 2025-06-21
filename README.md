# suggest

A CLI tool for suggesting shell commands using AI. Simply describe what you want to do, and `suggest` will provide the appropriate command tailored to your operating system.

## Setup

You must have one of the following API keys set in your environment:

### Option 1: OpenAI
```bash
export OPENAI_API_KEY="your-openai-api-key-here"
```

### Option 2: Anthropic
```bash
export ANTHROPIC_API_KEY="your-anthropic-api-key-here"
```

### Option 3: Google Gemini
```bash
export GEMINI_API_KEY="your-gemini-api-key-here"
```

To make these permanent, add the export line to your shell profile file (e.g., `~/.bashrc`, `~/.zshrc`, or `~/.profile`).

## Installation

```bash
go build -o suggest .
```

## Usage

```bash
./suggest "find all files larger than 1GB"
./suggest "compress all .log files in this directory"
./suggest "show me the last 10 commits"
```
