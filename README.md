# 🤖 gh-smart-commit

<div align="center">

** AI-Powered Git Assistant That Lives Entirely On Your Machine**

*Generate perfect commit messages, get code suggestions, and describe your work—all powered by local Ollama models*

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Privacy First](https://img.shields.io/badge/Privacy-First-green?style=flat&logo=shield)](https://github.com)
[![100% Local](https://img.shields.io/badge/100%25-Local-blue?style=flat&logo=home)](https://github.com)
[![No Data Tracking](https://img.shields.io/badge/No-Data%20Tracking-red?style=flat&logo=ghost)](https://github.com)

</div>

---

## ✨ Why gh-smart-commit?

> **Never write another boring commit message again.** Let AI handle the git ceremony while you focus on the code that matters.

🔒 **Privacy-First**: Your code never leaves your machine  
⚡ **Lightning Fast**: Local AI processing with Ollama  
🎯 **Intelligent**: Understands context, follows conventions  
🛠️ **Developer Friendly**: Integrates seamlessly with your workflow  

---

## 🎯 Features

| Feature | Status | Description |
|---------|--------|-------------|
| 🧠 **Smart Commit Messages** | ✅ **Ready** | Generate conventional commit messages from staged changes |
| 🔍 **Code Suggestions** | ✅ **Ready** | Get AI-powered improvement suggestions with severity levels |
| 📝 **Branch Descriptions** | ✅ **Ready** | Automatically describe what your branch accomplishes |
| 🏷️ **Tag Suggestions** | 🚧 **Coming Soon** | Get relevant tags/labels for your changes |

### 🛡️ Privacy & Security
- **100% Local Processing**: All AI happens on your machine via Ollama
- **Zero Cloud Dependencies**: No API keys, no external services
- **Optional Caching**: Smart caching only when you want it
- **Data Ownership**: Your code stays yours, always

---

##  Quick Start

### 📋 Prerequisites

```bash
# Install Go (1.21+ required)
brew install go

# Install Ollama - your local AI engine
brew install ollama

# Download a model (we recommend starting with llama3:8b)
ollama pull llama3:8b
```

### ⚡ Installation

```bash
# Clone and build
git clone <repository-url>
cd gh-smart-commit
go build -o gh-smart-commit .

# Or install directly (when published)
go install github.com/your-username/gh-smart-commit@latest
```

### 🎬 Your First AI Commit

```bash
# Make some changes to your code
echo "console.log('Hello AI!');" > hello.js

# Stage your changes
git add .

# Let AI write your commit message
gh-smart-commit smart-commit

# That's it! 🎉
```

---

## 🔧 Commands & Usage

### 🧠 `smart-commit` - Intelligent Commit Messages

*Transform your staged changes into perfect conventional commits*

```bash
gh-smart-commit smart-commit [flags]
```

**✨ What it does:**
- Analyzes your staged changes with context
- Generates conventional commit messages following best practices
- Streams responses for immediate feedback
- Validates message length and format
- Offers confirmation before committing

**🛠️ Flags:**
```bash
--auto-commit        Skip confirmation, commit immediately
--dry-run           Preview message without committing
--max-diff-lines    Limit diff analysis (default: 500)
```

**📖 Example:**
```bash
$ gh-smart-commit smart-commit

Generating commit message....

Generated commit message:
─────────────────────────
feat(auth): implement OAuth2 integration

Add Google and GitHub OAuth2 providers with secure token handling
and user profile synchronization.

- Configure OAuth2 client credentials
- Implement callback handlers
- Add user session management
─────────────────────────

Do you want to commit with this message? [y/N]: y
✓ Changes committed successfully!
```

---

### 🔍 `lint-suggestions` - Code Improvement Assistant

*Get AI-powered suggestions to make your code even better*

```bash
gh-smart-commit lint-suggestions [flags]
```

**✨ What it does:**
- Analyzes staged or unstaged changes
- Provides categorized improvement suggestions
- Color-codes suggestions by severity
- Respects NO_COLOR environment variable

**🎨 Severity Levels:**
- 🔴 **HIGH**: Critical issues that should be addressed
- 🟡 **MEDIUM**: Important improvements worth considering  
- 🟢 **LOW**: Nice-to-have enhancements

**🛠️ Flags:**
```bash
--staged            Analyze staged changes (default)
--unstaged          Analyze unstaged changes instead
--severity string   Filter by: all, high, medium, low (default: "all")
--max-suggestions   Limit suggestions shown (default: 10)
```

**📖 Example:**
```bash
$ gh-smart-commit lint-suggestions --severity high

Analyzing changes for improvement suggestions...

🔴 HIGH PRIORITY SUGGESTIONS:

1. Add error handling for database operations
   File: user.go:42
   Consider wrapping database calls with proper error handling

2. Potential memory leak in goroutine
   File: worker.go:15
   Goroutine may not terminate properly without context cancellation

🟡 MEDIUM PRIORITY SUGGESTIONS:

3. Consider using constants for magic numbers
   File: config.go:8
   Replace hardcoded values with named constants
```

---

### 📝 `branch-describe` - Branch Documentation

*Automatically document what your branch accomplishes*

```bash
gh-smart-commit branch-describe [flags]
```

**✨ What it does:**
- Analyzes recent commit history for context
- Generates comprehensive branch descriptions
- Perfect for PR descriptions and documentation
- Smart caching to avoid redundant API calls

**🛠️ Flags:**
```bash
--commits int       Commits to analyze (default: 10)
--no-cache         Skip cache, regenerate fresh
--base-branch      Compare against branch (default: "main")  
--include-stats    Show diff statistics (default: true)
```

**📖 Example:**
```bash
$ gh-smart-commit branch-describe

Analyzing branch history (10 commits)...

Branch Description:
─────────────────────────
This branch implements a comprehensive user authentication system with OAuth2 
integration for Google and GitHub providers. The implementation includes secure 
token handling, user profile synchronization, and session management.

Key Changes:
• OAuth2 client configuration and provider setup
• Secure callback handlers with CSRF protection  
• User session management with Redis backend
• Profile synchronization and data mapping
• Comprehensive error handling and logging

Statistics: 15 files changed, 847 additions, 23 deletions
─────────────────────────
```

---

### 🏷️ `tag-suggest` - Smart Tagging *(Coming Soon)*

*Get relevant tags and labels for your changes*

```bash
gh-smart-commit tag-suggest [flags]  # 🚧 In Development
```

---

## ⚙️ Configuration

### 📁 Configuration File

Create `~/.config/gh-smart-commit.yaml` for persistent settings:

```yaml
# 🤖 Ollama Configuration
ollama:
  host: "127.0.0.1:11434"
  model: "llama3:8b"          # or codellama:7b, mistral:7b
  temperature: 0.3             # 0.0 = focused, 1.0 = creative

# 🌍 Global Settings  
verbose: false

# 🧠 Smart Commit Rules
smart-commit:
  max-diff-lines: 500
  rules:
    - "Commit title max 72 chars"
    - "Use imperative mood"
    - "Follow Conventional Commits standard"

# 🔍 Lint Suggestions
lint-suggestions:
  severity: "all"              # all, high, medium, low
  max-suggestions: 10

# 📝 Branch Descriptions
branch-describe:
  commits: 10
  base-branch: "main"
  cache-ttl: "24h"            # Cache descriptions for 24 hours
```

### 🌿 Environment Variables

Override any setting with environment variables:

```bash
export GH_SMART_COMMIT_OLLAMA_HOST="127.0.0.1:11434"
export GH_SMART_COMMIT_OLLAMA_MODEL="codellama:7b"
export GH_SMART_COMMIT_OLLAMA_TEMPERATURE="0.2"
export GH_SMART_COMMIT_VERBOSE="true"
```

### 🚩 Global Flags

Available for all commands:

```bash
--config string         Custom config file path
--ollama-host string    Ollama server (default: "127.0.0.1:11434")
--model string          Model to use (default: "llama3:8b")
--temperature float     Creativity level 0.0-1.0 (default: 0.3)
--verbose              Enable detailed output
```

---

## 🎭 Real-World Examples

### 🔥 The "Friday Afternoon" Commit

```bash
# You: *frantically stages 47 files before weekend*
git add .

# Also you: *dreads writing commit message*
gh-smart-commit smart-commit

# AI: "feat(ui): implement responsive dashboard with dark mode
# 
# Complete redesign of user dashboard with mobile-first approach,
# including dark mode toggle, real-time notifications, and 
# improved accessibility features."

# You: 😍 *accepts immediately*
```

### 🐛 The Bug Hunt

```bash
# You found a nasty bug, fixed it, but explaining it feels impossible
gh-smart-commit smart-commit --dry-run

# Output: "fix(auth): resolve race condition in token refresh
#
# Fix race condition where concurrent requests could cause token
# refresh to fail intermittently by adding proper mutex locking
# around refresh operations."

# You: *mind blown* 🤯
```

### 🔍 Code Review Prep

```bash
# Before creating that PR
gh-smart-commit lint-suggestions --severity high

# Fix the critical issues
gh-smart-commit branch-describe

# Perfect PR description generated ✨
```

---

## 🏗️ Architecture

```
   🎨 CLI Layer (Cobra)
        ↓
   🧠 Domain Logic 
   (Git ops, Prompts, Validation)
        ↓  
   🤖 LLM Adapter
   (Ollama HTTP with streaming)
        ↓
   ⚙️ Config Layer
   (Viper: files, env, flags)
```

**Design Principles:**
- **🔌 Modular**: Each component has a single responsibility
- **🧪 Testable**: Comprehensive unit tests for all packages  
- ** Fast**: Streaming responses and smart caching
- **🛡️ Reliable**: Robust error handling and retry logic

---

## 🧪 Development

### 🏗️ Building

```bash
# Build for your platform
go build -o gh-smart-commit .

# Cross-compile for different platforms
GOOS=linux GOARCH=amd64 go build -o gh-smart-commit-linux .
GOOS=windows GOARCH=amd64 go build -o gh-smart-commit.exe .
```

### 🧪 Testing

```bash
# Run all tests
go test ./...

# With coverage report
go test -cover ./...

# Test specific packages
go test ./pkg/ollama -v
go test ./pkg/git -v
go test ./pkg/prompt -v
go test ./pkg/cache -v
```

### 🤝 Contributing

We'd love your help making gh-smart-commit even better!

1. 🍴 Fork the repository
2. 🌿 Create a feature branch (`git checkout -b amazing-feature`)
3. ✨ Make your changes
4. 🧪 Add tests for new functionality
5. ✅ Run `go test ./...` to ensure everything works
6. 📝 Update documentation if needed
7. 📬 Submit a pull request

**🎯 Areas where we'd love contributions:**
- Additional prompt templates and optimizations
- Support for more AI models and providers
- Enhanced caching strategies
- Better terminal UI and user experience
- Performance optimizations

---

## 🚨 Troubleshooting

### 🔌 Ollama Connection Issues

```bash
# Check if Ollama is running
ollama list

# Start Ollama service
ollama serve

# Test the connection
curl http://127.0.0.1:11434/api/tags
```

**💡 Common fixes:**
- Ensure Ollama is installed and running
- Check firewall settings for port 11434
- Verify the model is downloaded: `ollama pull llama3:8b`

### 🤖 Model Not Found

```bash
# List what's available locally
ollama list

# Download a recommended model
ollama pull llama3:8b        # Great all-rounder
ollama pull codellama:7b     # Code-specialized
ollama pull mistral:7b       # Fast and efficient
```

### 📁 Git Repository Issues

```bash
# Ensure you're in a Git repository
git status

# Initialize if needed
git init

# Stage some changes before using smart-commit
git add .
```

### 💾 Cache Issues

```bash
# Clear cache if needed
rm -rf .git/gh-smart-commit-cache/

# Or disable caching entirely
gh-smart-commit branch-describe --no-cache
```

---

## 📊 Performance & Models

| Model | Size | Speed | Quality | Best For |
|-------|------|-------|---------|----------|
| llama3:8b | 4.7GB | ⭐⭐⭐ | ⭐⭐⭐⭐ | Balanced performance |
| codellama:7b | 3.8GB | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | Code-specific tasks |
| mistral:7b | 4.1GB | ⭐⭐⭐⭐ | ⭐⭐⭐ | Fast responses |
| llama3:70b | 40GB | ⭐ | ⭐⭐⭐⭐⭐ | Maximum quality |

**💡 Recommendations:**
- **Development**: `codellama:7b` for code-focused tasks
- **General Use**: `llama3:8b` for best balance  
- **Speed**: `mistral:7b` for fastest responses
- **Quality**: `llama3:70b` if you have the hardware

---

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

---

<div align="center">

**Made with ❤️ for developers who care about their commit history**

*Star ⭐ this repo if gh-smart-commit made your day better!*

[![Built with Go](https://img.shields.io/badge/Built%20with-Go-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![Powered by Ollama](https://img.shields.io/badge/Powered%20by-Ollama-FF6B00?style=for-the-badge)](https://ollama.ai/)

</div> 
