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
	Repo        string
	Branch      string
	Diff        string
	Commits     []git.Commit
	Rules       []string
	MaxLength   int
	Style       string
	Description string      // For bash command descriptions
	SystemInfo  interface{} // For system context information
}

// SmartCommitTemplate is the prompt template for generating commit messages
var SmartCommitTemplate = Template{
	System: `You are an expert software engineer skilled in writing clear, descriptive commit messages.

CRITICAL INSTRUCTIONS:
- Your response must be ONLY the commit message itself
- NO explanations, NO additional text, NO context
- NO phrases like "Here is the commit message:" or "This commit message..."
- NO quotes around the message unless they are part of the actual commit message
- Just the raw commit message and nothing else

Requirements for the commit message:
1. Start with an action verb in imperative mood (Add, Remove, Fix, Update, Refactor, etc.)
2. Include specific file names or components where changes were made
3. Keep the first line under 72 characters
4. Be descriptive but concise
5. Use natural language, not conventional commit format
6. Focus on what was changed and where it was changed

EXAMPLE OUTPUT FORMAT:
Add OAuth2 integration to AuthService and UserController
Remove unused points status property in ThirdView
Fix null pointer error in user validation service
Update installation instructions in README
Refactor database queries in ProductRepository

REMEMBER: 
- Start with action verb (Add, Remove, Fix, Update, etc.)
- Include file/component names
- NO conventional commit format like "feat:" or "fix:"
- Output ONLY the commit message. No other text whatsoever.`,

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

// BashTemplate is the prompt template for generating bash commands
var BashTemplate = Template{
	System: `You are an expert system administrator and command-line specialist. Generate safe, efficient bash commands based on user descriptions and system context.

CRITICAL INSTRUCTIONS:
- Your response must be ONLY the bash command itself
- NO explanations, NO additional text, NO context
- NO phrases like "Here is the command:" or "You can use:"
- NO markdown code blocks or formatting
- Just the raw bash command and nothing else
- Ensure the command is safe and appropriate for the given context
- Use standard Unix/Linux tools when possible
- Be mindful of the operating system and available tools

SAFETY GUIDELINES:
- Avoid destructive operations without explicit user intent
- Use appropriate flags for safety (e.g., -i for interactive confirmations)
- Prefer relative paths when working within a project
- Use standard tools available on most systems

EXAMPLE OUTPUT FORMAT:
find . -name "*.go" -type f
ls -la | grep "^d"
tar -czf backup.tar.gz src/
grep -r "TODO" --include="*.js" .

REMEMBER: 
- Output ONLY the bash command
- No explanations or context
- Make it safe and appropriate for the system`,

	User: `Task: {{.Description}}

System Context:
- OS: {{.SystemInfo.OS}}
- Architecture: {{.SystemInfo.Arch}}
- Working Directory: {{.SystemInfo.WorkingDir}}
- Shell: {{.SystemInfo.Shell}}
- User: {{.SystemInfo.User}}
{{if .SystemInfo.IsGitRepo}}
- Git Repository: {{.SystemInfo.Repo}}
- Current Branch: {{.SystemInfo.Branch}}
{{end}}

Current Directory Structure:
{{.SystemInfo.FileTree}}

Generate the bash command:`,
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
			"bash":             BashTemplate,
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
	// Remove common AI prefixes and cleanup
	prefixes := []string{
		"Here is the commit message:",
		"Commit message:",
		"The commit message is:",
		"Here's the commit message:",
		"```",
	}

	cleaned := strings.TrimSpace(message)
	for _, prefix := range prefixes {
		if strings.HasPrefix(cleaned, prefix) {
			cleaned = strings.TrimSpace(strings.TrimPrefix(cleaned, prefix))
		}
	}

	// Remove trailing punctuation like quotes or backticks
	cleaned = strings.Trim(cleaned, "`\"'")

	return strings.TrimSpace(cleaned)
}

// SanitizeBashCommand cleans up a generated bash command
func SanitizeBashCommand(command string) string {
	// Remove common AI prefixes and cleanup
	prefixes := []string{
		"Here is the command:",
		"The command is:",
		"Here's the command:",
		"You can use:",
		"Try this:",
		"Run:",
		"Execute:",
		"```bash",
		"```sh",
		"```",
		"$",
		"# ",
	}

	cleaned := strings.TrimSpace(command)

	// Remove prefixes
	for _, prefix := range prefixes {
		if strings.HasPrefix(cleaned, prefix) {
			cleaned = strings.TrimSpace(strings.TrimPrefix(cleaned, prefix))
		}
	}

	// Remove markdown code block endings
	cleaned = strings.TrimSuffix(cleaned, "```")

	// If it's a multi-line response, take only the first line that looks like a command
	lines := strings.Split(cleaned, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "//") {
			return line
		}
	}

	return strings.TrimSpace(cleaned)
}
