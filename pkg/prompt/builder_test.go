package prompt

import (
	"strings"
	"testing"

	"gh-smart-commit/pkg/git"
)

func TestNewBuilder(t *testing.T) {
	builder := NewBuilder()
	if builder == nil {
		t.Fatal("NewBuilder returned nil")
	}

	if len(builder.templates) != 4 {
		t.Errorf("Expected 4 templates, got %d", len(builder.templates))
	}
}

func TestBuildSmartCommit(t *testing.T) {
	builder := NewBuilder()
	ctx := Context{
		Repo:   "test-repo",
		Branch: "main",
		Diff:   "diff --git a/file.go b/file.go\n+func test() {}",
		Rules:  []string{"Rule 1", "Rule 2"},
	}

	system, user, err := builder.Build("smart-commit", ctx)
	if err != nil {
		t.Errorf("Build failed: %v", err)
	}

	if system == "" {
		t.Error("System prompt is empty")
	}

	if user == "" {
		t.Error("User prompt is empty")
	}

	// Check that context variables are included
	if !strings.Contains(user, "test-repo") {
		t.Error("User prompt doesn't contain repo name")
	}

	if !strings.Contains(user, "main") {
		t.Error("User prompt doesn't contain branch name")
	}

	if !strings.Contains(user, "func test()") {
		t.Error("User prompt doesn't contain diff")
	}

	if !strings.Contains(user, "Rule 1") {
		t.Error("User prompt doesn't contain rules")
	}
}

func TestBuildNonExistentTemplate(t *testing.T) {
	builder := NewBuilder()
	ctx := Context{
		Repo:   "test-repo",
		Branch: "main",
		Diff:   "test diff",
	}

	_, _, err := builder.Build("non-existent", ctx)
	if err == nil {
		t.Error("Expected error for non-existent template")
	}

	if !strings.Contains(err.Error(), "template not found") {
		t.Errorf("Expected 'template not found' error, got: %v", err)
	}
}

func TestAddTemplate(t *testing.T) {
	builder := NewBuilder()
	customTemplate := Template{
		System: "Custom system prompt",
		User:   "Custom user prompt with {{.Repo}}",
	}

	builder.AddTemplate("custom", customTemplate)

	ctx := Context{
		Repo: "test-repo",
	}

	system, user, err := builder.Build("custom", ctx)
	if err != nil {
		t.Errorf("Build failed: %v", err)
	}

	if system != "Custom system prompt" {
		t.Errorf("Expected custom system prompt, got: %s", system)
	}

	if !strings.Contains(user, "test-repo") {
		t.Error("Custom template didn't process variables")
	}
}

func TestValidateCommitMessage(t *testing.T) {
	tests := []struct {
		message string
		wantErr bool
	}{
		{"feat: add new feature", false},
		{"fix(api): resolve authentication issue", false},
		{"docs: update README", false},
		{"", true}, // empty message
		{"this is a very long commit message that exceeds the 72 character limit for the first line", true}, // too long
		{"missing colon in conventional format", true}, // no colon
	}

	for _, tt := range tests {
		err := ValidateCommitMessage(tt.message)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateCommitMessage(%q) error = %v, wantErr %v", tt.message, err, tt.wantErr)
		}
	}
}

func TestSanitizeCommitMessage(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  feat: add feature  ", "feat: add feature"},
		{"Commit message: feat: add feature", "feat: add feature"},
		{"Here's the commit message: feat: add feature", "feat: add feature"},
		{`"feat: add feature"`, "feat: add feature"},
		{"`feat: add feature`", "feat: add feature"},
		{"The commit message is: feat: add feature", "feat: add feature"},
	}

	for _, tt := range tests {
		result := SanitizeCommitMessage(tt.input)
		if result != tt.expected {
			t.Errorf("SanitizeCommitMessage(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
} 