# gh-smart-commit

AI-powered Git assistant that uses local Ollama models to generate commit messages and provide code assistance, entirely offline.

## Features

- **Smart Commit Messages**: Generate conventional commit messages from staged changes
- **Code Suggestions**: Get AI-powered improvement suggestions for your code
- **Branch Descriptions**: Automatically describe what your branch accomplishes
- **Tag Suggestions**: Get relevant tags/labels for your changes
- **100% Local**: All AI processing happens locally via Ollama - no data leaves your machine
- **Privacy First**: Zero persistent storage of diffs unless you enable caching

## Quick Start

### Prerequisites

```bash
# Install Go (1.21+)
brew install go

# Install Ollama
brew install ollama

# Download a model (example)
ollama pull llama3:8b
```

### Installation

```bash
# Build from source
git clone <repository-url>
cd gh-smart-commit
go build -o gh-smart-commit .

# Or install directly
go install .
```

### Basic Usage

```bash
# Stage your changes
git add .

# Generate commit message
gh-smart-commit smart-commit

# Other commands
gh-smart-commit lint-suggestions     # Get code improvement suggestions
gh-smart-commit branch-describe      # Describe current branch
gh-smart-commit tag-suggest          # Suggest relevant tags
```

## Commands

### smart-commit

Generate conventional commit messages from staged changes.

```bash
gh-smart-commit smart-commit [flags]

Flags:
  --auto-commit        Automatically commit without confirmation
  --dry-run           Show generated message without committing
  --max-diff-lines    Maximum diff lines to include (default 500)
```

### lint-suggestions

Analyze code changes and suggest improvements.

```bash
gh-smart-commit lint-suggestions [flags]

Flags:
  --staged            Analyze staged changes (default)
  --unstaged          Analyze unstaged changes
  --severity string   Filter by severity: all, high, medium, low (default "all")
  --max-suggestions   Maximum suggestions to display (default 10)
```

### branch-describe

Generate a description of what your branch accomplishes.

```bash
gh-smart-commit branch-describe [flags]

Flags:
  --commits int        Number of recent commits to analyze (default 10)
  --no-cache          Skip cache and regenerate description
  --base-branch       Base branch to compare against (default "main")
  --include-stats     Include diff statistics (default true)
```

### tag-suggest

Suggest relevant tags or labels for your changes.

```bash
gh-smart-commit tag-suggest [flags]

Flags:
  --allowed-tags      Comma-separated list of allowed tags
  --max-tags int      Maximum number of tags to suggest (default 5)
  --validate-only     Only suggest from allowed tags list
  --include-auto      Include automatically detected tags (default true)
```

## Configuration

### Configuration File

Create `~/.config/gh-smart-commit.yaml`:

```yaml
# Ollama settings
ollama:
  host: "127.0.0.1:11434"
  model: "llama3:8b"
  temperature: 0.3

# Global settings
verbose: false

# Command-specific settings
smart-commit:
  max-diff-lines: 500
  rules:
    - "Commit title max 72 chars"
    - "Use imperative mood"
    - "Follow Conventional Commits standard"

lint-suggestions:
  severity: "all"
  max-suggestions: 10

branch-describe:
  commits: 10
  base-branch: "main"

tag-suggest:
  max-tags: 5
  allowed-tags:
    - "frontend"
    - "backend"
    - "api"
    - "ui"
    - "bugfix"
    - "feature"
    - "refactor"
    - "docs"
```

### Environment Variables

All configuration can be overridden with environment variables:

```bash
export GH_SMART_COMMIT_OLLAMA_HOST="127.0.0.1:11434"
export GH_SMART_COMMIT_OLLAMA_MODEL="llama3:8b"
export GH_SMART_COMMIT_OLLAMA_TEMPERATURE="0.3"
export GH_SMART_COMMIT_VERBOSE="true"
```

### Command Line Flags

```bash
# Global flags (available for all commands)
--config string         Config file path
--ollama-host string    Ollama server host:port (default "127.0.0.1:11434")
--model string          Ollama model to use (default "llama3:8b")
--temperature float     Model temperature 0.0-1.0 (default 0.3)
--verbose              Enable verbose output
```

## Examples

### Generate Commit Message

```bash
# Basic usage
git add .
gh-smart-commit smart-commit

# Output:
# Generating commit message....
# 
# Generated commit message:
# ─────────────────────────
# feat(cli): add smart commit message generation
# 
# Implement AI-powered commit message generation using Ollama
# with streaming support and user confirmation workflow.
# ─────────────────────────
# 
# Do you want to commit with this message? [y/N]: y
# ✓ Changes committed successfully!
```

### Dry Run Mode

```bash
gh-smart-commit smart-commit --dry-run

# Shows the generated message without committing
```

### Auto Commit

```bash
gh-smart-commit smart-commit --auto-commit

# Commits automatically without asking for confirmation
```

### Get Code Suggestions

```bash
gh-smart-commit lint-suggestions

# Output:
# Analyzing changes for improvement suggestions...
# 
# Suggestions for improvement:
# 
# 1. [HIGH] Add error handling
#    Function `processData` should handle potential nil pointer errors
# 
# 2. [MEDIUM] Use context for cancellation
#    Consider adding context.Context parameter for long-running operations
# 
# 3. [LOW] Add documentation
#    Public function lacks godoc comments
```

## Architecture

```
CLI Layer (Cobra)
    ↓
Domain Logic (Git operations, Prompt building)
    ↓
LLM Adapter (Ollama HTTP client with streaming)
    ↓
Config Layer (Viper - file, env, flags)
```

## Development

### Building

```bash
go build -o gh-smart-commit .
```

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./pkg/ollama
go test ./pkg/git
go test ./pkg/prompt
```

### Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run `go test ./...` to ensure tests pass
6. Submit a pull request

## Troubleshooting

### Ollama Connection Issues

```bash
# Check if Ollama is running
ollama list

# Start Ollama if not running
ollama serve

# Test connection
curl http://127.0.0.1:11434/api/tags
```

### Model Not Found

```bash
# List available models
ollama list

# Pull a model if needed
ollama pull llama3:8b
```

### Git Repository Issues

```bash
# Ensure you're in a Git repository
git status

# Stage some changes first
git add .
```

## License

MIT License - see [LICENSE](LICENSE) file for details. 