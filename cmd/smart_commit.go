package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"gh-smart-commit/pkg/git"
	"gh-smart-commit/pkg/ollama"
	"gh-smart-commit/pkg/prompt"
)

// smartCommitCmd represents the smart-commit command
var smartCommitCmd = &cobra.Command{
	Use:   "smart-commit",
	Short: "Generate commit messages from staged changes",
	Long: `Generate conventional commit messages by analyzing your staged Git changes.

This command reads the staged changes (git diff --cached) and uses AI to suggest
an appropriate commit message following Conventional Commits standard.

The suggested message will be under 72 characters for the first line and use
imperative mood as recommended by Git best practices.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSmartCommit(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(smartCommitCmd)

	// Command-specific flags
	smartCommitCmd.Flags().Bool("auto-commit", false, "Automatically commit with generated message (no confirmation)")
	smartCommitCmd.Flags().Bool("dry-run", false, "Show generated message without committing")
	smartCommitCmd.Flags().Int("max-diff-lines", 500, "Maximum diff lines to include in prompt")
}

func runSmartCommit(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Get flags
	autoCommit, _ := cmd.Flags().GetBool("auto-commit")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	maxDiffLines, _ := cmd.Flags().GetInt("max-diff-lines")
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

	// Get staged diff
	diff, err := repo.GetStagedDiff(ctx)
	if err != nil {
		return fmt.Errorf("failed to get staged diff: %w", err)
	}

	if strings.TrimSpace(diff) == "" {
		return fmt.Errorf("no staged changes found. Please stage your changes with 'git add' first")
	}

	// Truncate diff if too long
	if maxDiffLines > 0 {
		diff = git.TruncateDiff(diff, maxDiffLines)
	}

	// Get repository context
	repoName, _ := repo.GetRepoName(ctx)
	branch, _ := repo.GetCurrentBranch(ctx)

	if verbose {
		fmt.Fprintf(os.Stderr, "Repository: %s\n", repoName)
		fmt.Fprintf(os.Stderr, "Branch: %s\n", branch)
		fmt.Fprintf(os.Stderr, "Diff length: %d lines\n", len(strings.Split(diff, "\n")))
	}

	// Build prompt
	builder := prompt.NewBuilder()
	promptCtx := prompt.Context{
		Repo:   repoName,
		Branch: branch,
		Diff:   diff,
		Rules: []string{
			"Commit title max 72 chars",
			"Use imperative mood",
			"Follow Conventional Commits standard",
		},
	}

	systemPrompt, userPrompt, err := builder.Build("smart-commit", promptCtx)
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
	fmt.Print("Generating commit message")
	respChan, errChan := client.Chat(ctx, chatReq)

	var commitMessage strings.Builder
	var streamErr error

	for {
		select {
		case resp, ok := <-respChan:
			if !ok {
				// Channel closed, we're done
				goto StreamComplete
			}
			fmt.Print(".")
			commitMessage.WriteString(resp.Message.Content)

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
		return fmt.Errorf("failed to generate commit message: %w", streamErr)
	}

	// Clean up the generated message
	message := prompt.SanitizeCommitMessage(commitMessage.String())

	if message == "" {
		return fmt.Errorf("generated commit message is empty")
	}

	// Validate the message
	if err := prompt.ValidateCommitMessage(message); err != nil {
		fmt.Printf("Warning: %v\n", err)
	}

	// Display the generated message
	fmt.Printf("\nGenerated commit message:\n")
	fmt.Printf("─────────────────────────\n")
	fmt.Printf("%s\n", message)
	fmt.Printf("─────────────────────────\n")

	if dryRun {
		fmt.Println("\nDry run mode - not committing")
		return nil
	}

	// Ask for confirmation unless auto-commit is enabled
	if !autoCommit {
		fmt.Print("\nDo you want to commit with this message? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read user input: %w", err)
		}

		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Commit cancelled")
			return nil
		}
	}

	// Commit the changes
	if verbose {
		fmt.Fprintf(os.Stderr, "Committing changes...\n")
	}

	commitCmd := fmt.Sprintf(`git commit -m %q`, message)
	if err := runShellCommand(ctx, commitCmd); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	fmt.Println("✓ Changes committed successfully!")
	return nil
}

// runShellCommand executes a shell command
func runShellCommand(ctx context.Context, command string) error {
	args := []string{"-c", command}
	cmd := exec.CommandContext(ctx, "sh", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
