package ui

import (
	"fmt"
	"gh-smart-commit/pkg/git"
	"strings"
)

// CommitMessageFormatter handles formatting commit messages beautifully
type CommitMessageFormatter struct{}

// NewCommitMessageFormatter creates a new commit message formatter
func NewCommitMessageFormatter() *CommitMessageFormatter {
	return &CommitMessageFormatter{}
}

// FormatGenerated formats a generated commit message with beautiful styling
func (f *CommitMessageFormatter) FormatGenerated(message string) string {
	if IsNoColor() {
		return fmt.Sprintf(`
Generated commit message:
─────────────────────────
%s
─────────────────────────`, message)
	}

	header := HeaderStyle.Render("✨ Generated Commit Message")
	separator := CreateSeparator(60)
	messageStyled := CommitMessageStyle.Render(message)

	return fmt.Sprintf("\n%s\n%s\n%s\n%s\n",
		header,
		separator,
		messageStyled,
		separator)
}

// FormatConfirmation formats the confirmation prompt
func (f *CommitMessageFormatter) FormatConfirmation() string {
	if IsNoColor() {
		return "\nDo you want to commit with this message? [y/N]: "
	}

	prompt := InfoStyle.Render("Do you want to commit with this message?")
	options := MutedStyle.Render("[y/N]")

	return fmt.Sprintf("\n%s %s: ", prompt, options)
}

// BashCommandFormatter handles formatting bash commands beautifully
type BashCommandFormatter struct{}

// NewBashCommandFormatter creates a new bash command formatter
func NewBashCommandFormatter() *BashCommandFormatter {
	return &BashCommandFormatter{}
}

// FormatGenerated formats a generated bash command with beautiful styling
func (f *BashCommandFormatter) FormatGenerated(command string) string {
	if IsNoColor() {
		return fmt.Sprintf(`
─────────────────────────
%s
─────────────────────────`, command)
	}

	header := HeaderStyle.Render("Generated Bash Command")
	separator := CreateSeparator(60)
	commandStyled := CodeStyle.Render(command)

	return fmt.Sprintf("\n%s\n%s\n%s\n%s\n",
		header,
		separator,
		commandStyled,
		separator)
}

// FormatConfirmation formats the confirmation prompt for command execution
func (f *BashCommandFormatter) FormatConfirmation() string {
	if IsNoColor() {
		return "\nDo you want to execute this command? [y/N]: "
	}

	prompt := InfoStyle.Render("Do you want to execute this command?")
	options := MutedStyle.Render("[y/N]")

	return fmt.Sprintf("\n%s %s: ", prompt, options)
}

// SuggestionFormatter handles formatting lint suggestions
type SuggestionFormatter struct{}

// NewSuggestionFormatter creates a new suggestion formatter
func NewSuggestionFormatter() *SuggestionFormatter {
	return &SuggestionFormatter{}
}

// FormatSuggestionsList formats a list of suggestions with beautiful styling
func (f *SuggestionFormatter) FormatSuggestionsList(suggestions []Suggestion, diffType string, total int) string {
	if len(suggestions) == 0 {
		return RenderWarningBox("No suggestions found matching your criteria")
	}

	var result strings.Builder

	// Header
	header := fmt.Sprintf("💡 Code Improvement Suggestions (%s changes)", diffType)
	if IsNoColor() {
		result.WriteString(fmt.Sprintf("\n%s\n", header))
		result.WriteString(strings.Repeat("─", 60) + "\n\n")
	} else {
		result.WriteString("\n" + HeaderStyle.Render(header) + "\n")
		result.WriteString(CreateSeparator(60) + "\n\n")
	}

	// Suggestions
	for i, suggestion := range suggestions {
		result.WriteString(f.FormatSuggestion(i+1, suggestion))
		result.WriteString("\n")
	}

	// Footer
	result.WriteString(f.FormatSuggestionsSummary(len(suggestions), total))

	return result.String()
}

// FormatSuggestion formats a single suggestion
func (f *SuggestionFormatter) FormatSuggestion(number int, suggestion Suggestion) string {
	if IsNoColor() {
		return fmt.Sprintf("%d. [%s] %s\n   %s",
			number,
			suggestion.Severity,
			suggestion.Title,
			suggestion.Description)
	}

	icon := GetSeverityIcon(suggestion.Severity)
	severityStyle := GetSeverityStyle(suggestion.Severity)

	title := fmt.Sprintf("%s %s",
		severityStyle.Render(fmt.Sprintf("[%s]", suggestion.Severity)),
		BodyStyle.Render(suggestion.Title))

	var result strings.Builder
	result.WriteString(fmt.Sprintf("%s %d. %s\n", icon, number, title))

	if suggestion.Description != "" {
		description := MutedStyle.Render("   " + suggestion.Description)
		result.WriteString(description)
	}

	return result.String()
}

// FormatSuggestionsSummary formats the summary at the end
func (f *SuggestionFormatter) FormatSuggestionsSummary(shown, total int) string {
	if IsNoColor() {
		summary := fmt.Sprintf("Found %d suggestions", shown)
		if total > shown {
			summary += fmt.Sprintf(" (filtered from %d total)", total)
		}
		return "\n" + summary + "\n"
	}

	summary := fmt.Sprintf("Found %s suggestions",
		SuccessStyle.Render(fmt.Sprintf("%d", shown)))

	if total > shown {
		summary += MutedStyle.Render(fmt.Sprintf(" (filtered from %d total)", total))
	}

	return "\n" + InfoStyle.Render("💡 ") + summary + "\n"
}

// BranchFormatter handles formatting branch descriptions
type BranchFormatter struct{}

// NewBranchFormatter creates a new branch formatter
func NewBranchFormatter() *BranchFormatter {
	return &BranchFormatter{}
}

// FormatDescription formats a branch description beautifully
func (f *BranchFormatter) FormatDescription(description string, cached bool) string {
	if IsNoColor() {
		header := "Branch Description"
		if cached {
			header += " (cached)"
		}
		return fmt.Sprintf(`
%s:
─────────────────────
%s
`, header, description)
	}

	var header string
	if cached {
		header = HeaderStyle.Render("📄 Branch Description") +
			MutedStyle.Render(" (cached)")
	} else {
		header = HeaderStyle.Render("📄 Branch Description")
	}

	separator := CreateSeparator(60)
	content := BodyStyle.Render(description)

	result := fmt.Sprintf("\n%s\n%s\n%s\n",
		header,
		separator,
		content)

	if cached {
		cacheNote := MutedStyle.Render("💾 From cache • Use --no-cache to regenerate")
		result += "\n" + cacheNote + "\n"
	}

	return result
}

// FormatStats formats diff statistics
func (f *BranchFormatter) FormatStats(stats string) string {
	if IsNoColor() {
		return "\nStatistics: " + stats + "\n"
	}

	return "\n" + InfoStyle.Render("📊 Statistics: ") +
		MutedStyle.Render(stats) + "\n"
}

// Suggestion represents a code improvement suggestion
type Suggestion struct {
	Severity    string
	Title       string
	Description string
	Number      int
}

// ContextFormatter handles formatting context information
type ContextFormatter struct{}

// NewContextFormatter creates a new context formatter
func NewContextFormatter() *ContextFormatter {
	return &ContextFormatter{}
}

// FormatRepoInfo formats repository information
func (f *ContextFormatter) FormatRepoInfo(repoName, branch string, verbose bool) string {
	if !verbose {
		return ""
	}

	if IsNoColor() {
		return fmt.Sprintf("Repository: %s\nBranch: %s\n", repoName, branch)
	}

	repo := MutedStyle.Render("Repository: ") + InfoStyle.Render(repoName)
	branchInfo := MutedStyle.Render("Branch: ") + InfoStyle.Render(branch)

	return fmt.Sprintf("%s\n%s\n", repo, branchInfo)
}

// FormatCommitList formats a list of commits
func (f *ContextFormatter) FormatCommitList(commits []git.Commit) string {
	if len(commits) == 0 {
		return ""
	}

	var result strings.Builder

	if IsNoColor() {
		result.WriteString("Recent commits:\n")
		for i, commit := range commits {
			if i >= 5 { // Limit display
				break
			}
			result.WriteString(fmt.Sprintf("  • %s\n", commit.Message))
		}
	} else {
		result.WriteString(MutedStyle.Render("Recent commits:") + "\n")
		for i, commit := range commits {
			if i >= 5 { // Limit display
				break
			}
			result.WriteString(MutedStyle.Render("  • ") +
				BodyStyle.Render(commit.Message) + "\n")
		}
	}

	return result.String()
}
