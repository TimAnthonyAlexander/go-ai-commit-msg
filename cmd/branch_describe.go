package cmd

import (
	"context"
	"fmt"
	"gh-smart-commit/pkg/cache"
	"gh-smart-commit/pkg/git"
	"gh-smart-commit/pkg/ollama"
	"gh-smart-commit/pkg/prompt"
	"os"
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
		return fmt.Errorf("failed to check if inside Git repository: %w", err)
	}
	if !isGit {
		return fmt.Errorf("not inside a Git repository")
	}

	// Get repository context
	repoName, _ := repo.GetRepoName(ctx)
	currentBranch, err := repo.GetCurrentBranch(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Repository: %s\n", repoName)
		fmt.Fprintf(os.Stderr, "Current branch: %s\n", currentBranch)
		fmt.Fprintf(os.Stderr, "Base branch: %s\n", baseBranch)
		fmt.Fprintf(os.Stderr, "Commits to analyze: %d\n", commitCount)
	}

	// Initialize cache
	cacheInstance := cache.NewCache(".")
	cacheKey := cache.GenerateCacheKey(
		"branch-describe",
		repoName,
		currentBranch,
		baseBranch,
		fmt.Sprintf("commits-%d", commitCount),
		fmt.Sprintf("stats-%t", includeStats),
	)

	// Check cache first (unless --no-cache is specified)
	if !noCache {
		if cachedDescription, found, err := cacheInstance.Get(cacheKey); err == nil && found {
			if verbose {
				fmt.Fprintf(os.Stderr, "Using cached description\n")
			}

			fmt.Printf("📄 Branch Description (cached):\n")
			fmt.Println("──────────────────────────────")
			fmt.Printf("%s\n", cachedDescription)
			fmt.Printf("\n💾 From cache • Use --no-cache to regenerate\n")
			return nil
		} else if err != nil && verbose {
			fmt.Fprintf(os.Stderr, "Cache error (proceeding without cache): %v\n", err)
		}
	}

	// Get recent commits
	commits, err := repo.GetRecentCommits(ctx, commitCount)
	if err != nil {
		return fmt.Errorf("failed to get recent commits: %w", err)
	}

	if len(commits) == 0 {
		return fmt.Errorf("no commits found on branch %s", currentBranch)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Found %d commits\n", len(commits))
	}

	// Get branch comparison diff if base branch is specified
	var branchDiff string
	if baseBranch != "" && baseBranch != currentBranch {
		// Try to get diff against base branch
		if diff, diffErr := getBranchDiff(ctx, repo, baseBranch, currentBranch); diffErr == nil {
			branchDiff = diff
			if verbose {
				fmt.Fprintf(os.Stderr, "Branch diff length: %d lines\n", len(strings.Split(branchDiff, "\n")))
			}
		} else if verbose {
			fmt.Fprintf(os.Stderr, "Could not get branch diff: %v\n", diffErr)
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
		return fmt.Errorf("failed to build prompt: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Sending request to Ollama...\n")
	}

	// Create Ollama client
	ollamaHost := viper.GetString("ollama.host")
	if !strings.HasPrefix(ollamaHost, "http") {
		ollamaHost = "http://" + ollamaHost
	}

	client := ollama.NewClient(ollamaHost)

	// Test connection
	if err := client.Ping(ctx); err != nil {
		return fmt.Errorf("failed to connect to Ollama at %s: %w", ollamaHost, err)
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

	// Stream response
	fmt.Print("Generating branch description")
	respChan, errChan := client.Chat(ctx, chatReq)

	var responseBuilder strings.Builder
	var streamErr error

	for {
		select {
		case resp, ok := <-respChan:
			if !ok {
				goto StreamComplete
			}
			fmt.Print(".")
			responseBuilder.WriteString(resp.Message.Content)

		case err := <-errChan:
			streamErr = err
			goto StreamComplete

		case <-ctx.Done():
			return ctx.Err()
		}
	}

StreamComplete:
	fmt.Println() // New line after dots

	if streamErr != nil {
		return fmt.Errorf("failed to generate branch description: %w", streamErr)
	}

	description := strings.TrimSpace(responseBuilder.String())
	if description == "" {
		return fmt.Errorf("no description generated")
	}

	// Clean up the description
	description = cleanupDescription(description)

	// Cache the result (expire after 24 hours)
	if !noCache {
		if err := cacheInstance.Set(cacheKey, description, 24*time.Hour); err != nil && verbose {
			fmt.Fprintf(os.Stderr, "Failed to cache result: %v\n", err)
		}
	}

	// Display the description
	fmt.Printf("\n📄 Branch Description:\n")
	fmt.Println("─────────────────────")
	fmt.Printf("%s\n", description)

	// Show summary stats
	if includeStats && len(commits) > 0 {
		totalFiles := 0
		totalAdditions := 0
		totalDeletions := 0

		for _, commit := range commits {
			totalFiles += len(commit.Files)
			totalAdditions += commit.Additions
			totalDeletions += commit.Deletions
		}

		fmt.Printf("\n📊 Branch Statistics:\n")
		fmt.Printf("• %d commits analyzed\n", len(commits))
		fmt.Printf("• %d files changed\n", totalFiles)
		fmt.Printf("• +%d additions, -%d deletions\n", totalAdditions, totalDeletions)
	}

	if !noCache {
		fmt.Printf("\n💾 Cached for 24 hours • Use --no-cache to regenerate\n")
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
