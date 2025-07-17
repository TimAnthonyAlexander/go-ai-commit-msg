package git

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Repository represents a Git repository interface
type Repository interface {
	GetStagedDiff(ctx context.Context) (string, error)
	GetUnstagedDiff(ctx context.Context) (string, error)
	GetCurrentBranch(ctx context.Context) (string, error)
	GetRepoName(ctx context.Context) (string, error)
	GetRecentCommits(ctx context.Context, count int) ([]Commit, error)
	IsInsideWorkTree(ctx context.Context) (bool, error)
}

// Commit represents a Git commit
type Commit struct {
	Hash      string
	Message   string
	Author    string
	Date      string
	Files     []string
	Additions int
	Deletions int
}

// LocalRepo implements Repository for local Git repositories
type LocalRepo struct {
	workDir string
}

// NewLocalRepo creates a new local repository instance
func NewLocalRepo(workDir string) *LocalRepo {
	if workDir == "" {
		workDir = "."
	}
	return &LocalRepo{workDir: workDir}
}

// GetStagedDiff returns the staged changes
func (r *LocalRepo) GetStagedDiff(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "diff", "--cached", "--no-pager")
	cmd.Dir = r.workDir

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get staged diff: %w", err)
	}

	return string(output), nil
}

// GetUnstagedDiff returns the unstaged changes
func (r *LocalRepo) GetUnstagedDiff(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "diff", "--no-pager")
	cmd.Dir = r.workDir

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get unstaged diff: %w", err)
	}

	return string(output), nil
}

// GetCurrentBranch returns the current branch name
func (r *LocalRepo) GetCurrentBranch(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "branch", "--show-current")
	cmd.Dir = r.workDir

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// GetRepoName returns the repository name
func (r *LocalRepo) GetRepoName(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "remote", "get-url", "origin")
	cmd.Dir = r.workDir

	output, err := cmd.Output()
	if err != nil {
		// Fallback to directory name if no remote
		cmd = exec.CommandContext(ctx, "basename", r.workDir)
		output, err = cmd.Output()
		if err != nil {
			return "unknown", nil
		}
	}

	repoURL := strings.TrimSpace(string(output))

	// Extract repo name from URL
	parts := strings.Split(repoURL, "/")
	if len(parts) > 0 {
		name := parts[len(parts)-1]
		name = strings.TrimSuffix(name, ".git")
		return name, nil
	}

	return "unknown", nil
}

// GetRecentCommits returns recent commits with statistics
func (r *LocalRepo) GetRecentCommits(ctx context.Context, count int) ([]Commit, error) {
	// Get commit info
	cmd := exec.CommandContext(ctx, "git", "log",
		fmt.Sprintf("-%d", count),
		"--pretty=format:%H|%s|%an|%ad",
		"--date=short",
	)
	cmd.Dir = r.workDir

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get recent commits: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	commits := make([]Commit, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) != 4 {
			continue
		}

		commit := Commit{
			Hash:    parts[0],
			Message: parts[1],
			Author:  parts[2],
			Date:    parts[3],
		}

		// Get file stats for this commit
		statsCmd := exec.CommandContext(ctx, "git", "show", "--stat", "--format=", commit.Hash)
		statsCmd.Dir = r.workDir

		statsOutput, err := statsCmd.Output()
		if err == nil {
			commit.Files, commit.Additions, commit.Deletions = parseGitStats(string(statsOutput))
		}

		commits = append(commits, commit)
	}

	return commits, nil
}

// IsInsideWorkTree checks if we're inside a Git repository
func (r *LocalRepo) IsInsideWorkTree(ctx context.Context) (bool, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = r.workDir

	output, err := cmd.Output()
	if err != nil {
		return false, nil // Not a Git repo
	}

	return strings.TrimSpace(string(output)) == "true", nil
}

// parseGitStats parses git --stat output to extract file changes and line counts
func parseGitStats(stats string) (files []string, additions, deletions int) {
	lines := strings.Split(stats, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Skip summary line
		if strings.Contains(line, "file") && (strings.Contains(line, "changed") || strings.Contains(line, "insertion") || strings.Contains(line, "deletion")) {
			continue
		}

		// Extract filename (before the first |)
		if strings.Contains(line, "|") {
			parts := strings.Split(line, "|")
			if len(parts) >= 2 {
				filename := strings.TrimSpace(parts[0])
				if filename != "" {
					files = append(files, filename)
				}

				// Count + and - in the second part
				changes := parts[1]
				additions += strings.Count(changes, "+")
				deletions += strings.Count(changes, "-")
			}
		}
	}

	return files, additions, deletions
}

// TruncateDiff truncates a diff to a maximum number of lines
func TruncateDiff(diff string, maxLines int) string {
	if maxLines <= 0 {
		return diff
	}

	lines := strings.Split(diff, "\n")
	if len(lines) <= maxLines {
		return diff
	}

	truncated := strings.Join(lines[:maxLines], "\n")
	truncated += fmt.Sprintf("\n\n...(diff truncated after %d lines)", maxLines)

	return truncated
}
