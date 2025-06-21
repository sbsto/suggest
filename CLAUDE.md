# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go project named `suggest` targeting Go 1.24.4. The project is currently minimal with only a `go.mod` file.

## Development Commands

```bash
# Build the CLI tool
go build -o suggest .

# Run tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run a specific test
go test -run TestName ./...

# Format code
go fmt ./...

# Vet code for issues
go vet ./...

# Install/update dependencies
go mod tidy

# Test the CLI locally
./suggest "find all files larger than 1GB"
```

## Architecture

This is a CLI tool built with Cobra that suggests shell commands using AI APIs:

- **main.go**: Single-file CLI application with Cobra command structure
- **API Integration**: Supports OpenAI, Anthropic, and Gemini APIs (checks for keys in that order)
- **Interactive Flow**: After getting suggestion, user can press Enter to run, 'y' to copy, or exit
- **Dependencies**: Uses Cobra for CLI, clipboard package for copy functionality, and respective API clients

The tool checks for API keys in environment variables:
- `OPENAI_API_KEY` (checked first)
- `ANTHROPIC_API_KEY` (checked second) 
- `GEMINI_API_KEY` (checked third)