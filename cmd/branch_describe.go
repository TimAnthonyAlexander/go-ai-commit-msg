package cmd

import (
	"context"
	"fmt"
	"gh-smart-commit/pkg/cache"
	"gh-smart-commit/pkg/git"
	"gh-smart-commit/pkg/ollama"
	"gh-smart-commit/pkg/prompt"
	"gh-smart-commit/pkg/ui"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// branchDescribeCmd represents the branch-describe command
var branchDescribeCmd = &cobra.Command{
	Use:   "branch-describe",
	Short: "Generate a description of the current branch's changes",
	Long: `Analyze the recent commits and changes in the current branch to generate
a concise description of what the branch accomplishes.

This is useful for:
- Pull request descriptions
- Release notes
- Branch documentation
- Code review preparation

Results are cached in .git/gh-smart-commit.cache to avoid repeated analysis.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBranchDescribe(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(branchDescribeCmd)

	// Command-specific flags
	branchDescribeCmd.Flags().Int("commits", 10, "Number of recent commits to analyze")
	branchDescribeCmd.Flags().Bool("no-cache", false, "Skip cache and regenerate description")
	branchDescribeCmd.Flags().String("base-branch", "main", "Base branch to compare against")
	branchDescribeCmd.Flags().Bool("include-stats", true, "Include diff statistics in analysis")
}

func runBranchDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Get flags
	commitCount, _ := cmd.Flags().GetInt("commits")
	noCache, _ := cmd.Flags().GetBool("no-cache")
	baseBranch, _ := cmd.Flags().GetString("base-branch")
	includeStats, _ := cmd.Flags().GetBool("include-stats")
	verbose := viper.GetBool("verbose")

	// Initialize Git repository
	repo := git.NewLocalRepo(".")

	// Check if we're in a Git repository
	isGit, err := repo.IsInsideWorkTree(ctx)
	if err != nil {
		ui.ShowError("Failed to check if inside Git repository: " + err.Error())
		return err
	}
	if !isGit {
		ui.ShowError("Not inside a Git repository")
		return fmt.Errorf("not inside a Git repository")
	}

	// Get repository context
	repoName, _ := repo.GetRepoName(ctx)
	currentBranch, _ := repo.GetCurrentBranch(ctx)

	// Show context info if verbose
	contextFormatter := ui.NewContextFormatter()
	if info := contextFormatter.FormatRepoInfo(repoName, currentBranch, verbose); info != "" {
		fmt.Print(info)
	}

	if verbose {
		ui.ShowInfo(fmt.Sprintf("Analyzing %d recent commits", commitCount))
		if baseBranch != "" && baseBranch != currentBranch {
			ui.ShowInfo(fmt.Sprintf("Comparing against base branch: %s", baseBranch))
		}
	}

	// Set up cache
	cacheInstance := cache.NewCache(".")
	cacheKey := fmt.Sprintf("branch-describe-%s-%d", currentBranch, commitCount)

	// Try to get from cache first
	if !noCache {
		if cachedDescription, found, err := cacheInstance.Get(cacheKey); err == nil && found {
			if verbose {
				ui.ShowInfo("Using cached description")
			}

			formatter := ui.NewBranchFormatter()
			output := formatter.FormatDescription(cachedDescription, true)
			fmt.Print(output)
			return nil
		} else if err != nil && verbose {
			ui.ShowInfo("Cache unavailable, generating fresh description")
		}
	}

	// Get recent commits
	commits, err := repo.GetRecentCommits(ctx, commitCount)
	if err != nil {
		ui.ShowError("Failed to get recent commits: " + err.Error())
		return err
	}

	if len(commits) == 0 {
		ui.ShowWarning(fmt.Sprintf("No commits found on branch %s", currentBranch))
		return fmt.Errorf("no commits found on branch %s", currentBranch)
	}

	if verbose {
		ui.ShowInfo(fmt.Sprintf("Found %d commits", len(commits)))

		// Show recent commits if very verbose
		contextFormatter := ui.NewContextFormatter()
		if commitInfo := contextFormatter.FormatCommitList(commits); commitInfo != "" {
			fmt.Print(commitInfo)
		}
	}

	// Get branch comparison diff if base branch is specified
	var branchDiff string
	if baseBranch != "" && baseBranch != currentBranch {
		// Try to get diff against base branch
		if diff, diffErr := getBranchDiff(ctx, repo, baseBranch, currentBranch); diffErr == nil {
			branchDiff = diff
			if verbose {
				diffLines := len(strings.Split(branchDiff, "\n"))
				ui.ShowInfo(fmt.Sprintf("Branch diff: %d lines", diffLines))
			}
		} else if verbose {
			ui.ShowWarning("Could not get branch diff: " + diffErr.Error())
		}
	}

	// Build prompt context
	builder := prompt.NewBuilder()
	promptCtx := prompt.Context{
		Repo:    repoName,
		Branch:  currentBranch,
		Commits: commits,
		Diff:    branchDiff,
	}

	systemPrompt, userPrompt, err := builder.Build("branch-describe", promptCtx)
	if err != nil {
		ui.ShowError("Failed to build prompt: " + err.Error())
		return err
	}

	if verbose {
		ui.ShowInfo("Sending request to Ollama...")
	}

	// Create Ollama client
	ollamaHost := viper.GetString("ollama.host")
	if !strings.HasPrefix(ollamaHost, "http") {
		ollamaHost = "http://" + ollamaHost
	}

	client := ollama.NewClient(ollamaHost)

	// Test connection
	if err := client.Ping(ctx); err != nil {
		ui.ShowError(fmt.Sprintf("Failed to connect to Ollama at %s: %s", ollamaHost, err.Error()))
		return err
	}

	// Prepare chat request
	chatReq := ollama.ChatRequest{
		Model: viper.GetString("ollama.model"),
		Messages: []ollama.Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Options: ollama.Options{
			Temperature: float32(viper.GetFloat64("ollama.temperature")),
		},
	}

	// Create beautiful streaming spinner
	spinner := ui.NewStreamingSpinner("📝 Generating branch description")
	spinner.Start()

	respChan, errChan := client.Chat(ctx, chatReq)

	var responseBuilder strings.Builder
	var streamErr error

	for {
		select {
		case resp, ok := <-respChan:
			if !ok {
				goto StreamComplete
			}
			spinner.Update()
			responseBuilder.WriteString(resp.Message.Content)

		case err := <-errChan:
			streamErr = err
			goto StreamComplete

		case <-ctx.Done():
			return ctx.Err()
		}
	}

StreamComplete:
	spinner.Stop()

	if streamErr != nil {
		ui.ShowError("Failed to generate branch description: " + streamErr.Error())
		return streamErr
	}

	description := strings.TrimSpace(responseBuilder.String())
	if description == "" {
		ui.ShowWarning("No description generated")
		return fmt.Errorf("no description generated")
	}

	// Clean up the description
	description = cleanupDescription(description)

	// Cache the result (expire after 24 hours)
	if !noCache {
		if err := cacheInstance.Set(cacheKey, description, 24*time.Hour); err != nil && verbose {
			ui.ShowWarning("Failed to cache result: " + err.Error())
		}
	}

	// Display the description beautifully
	formatter := ui.NewBranchFormatter()
	output := formatter.FormatDescription(description, false)
	fmt.Print(output)

	// Show summary stats if requested
	if includeStats {
		if stats := getStatsString(ctx, repo, baseBranch, currentBranch); stats != "" {
			statsOutput := formatter.FormatStats(stats)
			fmt.Print(statsOutput)
		}
	}

	return nil
}

// getBranchDiff gets the diff between two branches
func getBranchDiff(ctx context.Context, repo *git.LocalRepo, baseBranch, targetBranch string) (string, error) {
	// For now, we'll use a simple approach - this could be enhanced to use git diff branch..branch
	// But since our git package doesn't have this yet, we'll skip it for this implementation
	return "", fmt.Errorf("branch diff not implemented yet")
}

// cleanupDescription cleans up the AI-generated description
func cleanupDescription(description string) string {
	// Remove common AI prefixes
	prefixes := []string{
		"This branch", "The branch", "Branch description:", "Description:",
		"Here's a description:", "Here is a description:",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(description, prefix) {
			description = strings.TrimSpace(strings.TrimPrefix(description, prefix))
			description = strings.TrimPrefix(description, ":")
			description = strings.TrimSpace(description)
		}
	}

	// Ensure it starts with a capital letter
	if len(description) > 0 {
		description = strings.ToUpper(string(description[0])) + description[1:]
	}

	return description
}

// getStatsString generates statistics string for the branch
func getStatsString(ctx context.Context, repo *git.LocalRepo, baseBranch, currentBranch string) string {
	// Try to get some basic stats - this is a simplified implementation
	commits, err := repo.GetRecentCommits(ctx, 20) // Get more commits for stats
	if err != nil || len(commits) == 0 {
		return ""
	}

	totalFiles := 0
	totalAdditions := 0
	totalDeletions := 0

	for _, commit := range commits {
		totalFiles += len(commit.Files)
		totalAdditions += commit.Additions
		totalDeletions += commit.Deletions
	}

	return fmt.Sprintf("%d commits, %d files changed, +%d/-%d lines",
		len(commits), totalFiles, totalAdditions, totalDeletions)
}
