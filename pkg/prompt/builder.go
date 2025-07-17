package prompt

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"gh-smart-commit/pkg/git"
)

// Template represents a prompt template
type Template struct {
	System string
	User   string
}

// Context holds the context data for prompt templates
type Context struct {
	Repo      string
	Branch    string
	Diff      string
	Commits   []git.Commit
	Rules     []string
	MaxLength int
	Style     string
}

// SmartCommitTemplate is the prompt template for generating commit messages
var SmartCommitTemplate = Template{
	System: `You are an expert software engineer skilled in writing clear, concise commit messages following the Conventional Commits standard.

CRITICAL INSTRUCTIONS:
- Your response must be ONLY the commit message itself
- NO explanations, NO additional text, NO context
- NO phrases like "Here is the commit message:" or "This commit message..."
- NO quotes around the message unless they are part of the actual commit message
- Just the raw commit message and nothing else

Requirements for the commit message:
1. Follow Conventional Commits format: type(scope): description
2. Use imperative mood (e.g., "add", "fix", "update", not "added", "fixed", "updated")
3. Keep the first line under 72 characters
4. Use appropriate types: feat, fix, docs, style, refactor, test, chore, etc.
5. Include scope when relevant (e.g., api, ui, auth, db)
6. Be descriptive but concise

EXAMPLE OUTPUT FORMAT:
feat(auth): add OAuth2 integration with Google
fix(api): resolve null pointer error in user validation
docs: update installation instructions
refactor(db): optimize query performance

REMEMBER: Output ONLY the commit message. No other text whatsoever.`,

	User: `Repository: {{.Repo}}
Branch: {{.Branch}}

{{if .Rules}}Rules:
{{range .Rules}}- {{.}}
{{end}}
{{end}}

Diff:
{{.Diff}}

Output the commit message only:`,
}

// LintSuggestionsTemplate is the prompt template for code improvement suggestions
var LintSuggestionsTemplate = Template{
	System: `You are an expert code reviewer and software engineer. Analyze the provided code changes and suggest improvements focusing on:

1. Code quality and maintainability
2. Performance optimizations
3. Security considerations
4. Best practice adherence
5. Potential bugs or issues

Format your response as a numbered list where each suggestion includes:
- Severity level: [HIGH/MEDIUM/LOW]
- Brief description
- Specific recommendation

Keep suggestions actionable and specific. Focus on the most impactful improvements first.`,

	User: `Repository: {{.Repo}}
Branch: {{.Branch}}

Changes to review:
{{.Diff}}

Provide ordered suggestions for improvement:`,
}

// BranchDescribeTemplate is the prompt template for describing branch changes
var BranchDescribeTemplate = Template{
	System: `You are an expert software engineer who creates clear, concise descriptions of code changes for documentation purposes.

Analyze the provided commits and changes to generate a brief description (2-3 sentences) that explains:
1. What the branch accomplishes
2. The main changes or features implemented
3. The overall impact or purpose

Write in present tense and focus on the "what" and "why" rather than implementation details.`,

	User: `Repository: {{.Repo}}
Branch: {{.Branch}}

Recent commits:
{{range .Commits}}- {{.Message}} ({{.Date}})
{{end}}

{{if .Diff}}Recent changes:
{{.Diff}}
{{end}}

Generate a concise description of what this branch accomplishes:`,
}

// TagSuggestTemplate is the prompt template for suggesting tags
var TagSuggestTemplate = Template{
	System: `You are an expert at categorizing and tagging code changes. Analyze the provided changes and suggest relevant tags or labels.

Consider these aspects:
1. File types and languages involved
2. Areas of the codebase affected (frontend, backend, database, etc.)
3. Type of changes (feature, bugfix, refactor, performance, etc.)
4. Impact level (breaking, major, minor, patch)
5. Functional areas (authentication, ui, api, documentation, etc.)

Suggest 3-5 most relevant tags. Format as a simple comma-separated list.`,

	User: `Repository: {{.Repo}}
Branch: {{.Branch}}

Files changed:
{{range .Commits}}{{range .Files}}- {{.}}
{{end}}{{end}}

Changes:
{{.Diff}}

Suggest relevant tags (comma-separated):`,
}

// Builder builds prompts from templates and context
type Builder struct {
	templates map[string]Template
}

// NewBuilder creates a new prompt builder
func NewBuilder() *Builder {
	return &Builder{
		templates: map[string]Template{
			"smart-commit":     SmartCommitTemplate,
			"lint-suggestions": LintSuggestionsTemplate,
			"branch-describe":  BranchDescribeTemplate,
			"tag-suggest":      TagSuggestTemplate,
		},
	}
}

// Build builds a prompt for the given template name and context
func (b *Builder) Build(templateName string, ctx Context) (system, user string, err error) {
	tmpl, exists := b.templates[templateName]
	if !exists {
		return "", "", fmt.Errorf("template not found: %s", templateName)
	}

	// Build system prompt
	systemTmpl, err := template.New("system").Parse(tmpl.System)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse system template: %w", err)
	}

	var systemBuf bytes.Buffer
	if err := systemTmpl.Execute(&systemBuf, ctx); err != nil {
		return "", "", fmt.Errorf("failed to execute system template: %w", err)
	}

	// Build user prompt
	userTmpl, err := template.New("user").Parse(tmpl.User)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse user template: %w", err)
	}

	var userBuf bytes.Buffer
	if err := userTmpl.Execute(&userBuf, ctx); err != nil {
		return "", "", fmt.Errorf("failed to execute user template: %w", err)
	}

	return systemBuf.String(), userBuf.String(), nil
}

// AddTemplate adds a custom template
func (b *Builder) AddTemplate(name string, tmpl Template) {
	b.templates[name] = tmpl
}

// ValidateCommitMessage validates a generated commit message
func ValidateCommitMessage(message string) error {
	if message == "" {
		return fmt.Errorf("commit message is empty")
	}

	lines := strings.Split(message, "\n")
	if len(lines) == 0 {
		return fmt.Errorf("commit message is empty")
	}

	firstLine := strings.TrimSpace(lines[0])
	if len(firstLine) > 72 {
		return fmt.Errorf("first line is too long (%d chars, max 72)", len(firstLine))
	}

	// Basic conventional commit format check
	if !strings.Contains(firstLine, ":") {
		return fmt.Errorf("commit message should follow 'type: description' format")
	}

	return nil
}

// SanitizeCommitMessage cleans up a generated commit message
func SanitizeCommitMessage(message string) string {
	// Remove leading/trailing whitespace
	message = strings.TrimSpace(message)

	// Remove common prefixes that LLMs might add
	prefixes := []string{"Commit message:", "Here's the commit message:", "The commit message is:"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(message, prefix) {
			message = strings.TrimSpace(strings.TrimPrefix(message, prefix))
		}
	}

	// Remove quotes if the entire message is quoted
	if (strings.HasPrefix(message, `"`) && strings.HasSuffix(message, `"`)) ||
		(strings.HasPrefix(message, "`") && strings.HasSuffix(message, "`")) {
		message = message[1 : len(message)-1]
	}

	return strings.TrimSpace(message)
}
