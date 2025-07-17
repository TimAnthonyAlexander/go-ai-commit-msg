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
	"gh-smart-commit/pkg/ui"
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
		ui.ShowError("Failed to check if inside Git repository: " + err.Error())
		return err
	}
	if !isGit {
		ui.ShowError("Not inside a Git repository")
		return fmt.Errorf("not inside a Git repository")
	}

	// Get staged diff
	diff, err := repo.GetStagedDiff(ctx)
	if err != nil {
		ui.ShowError("Failed to get staged diff: " + err.Error())
		return err
	}

	if strings.TrimSpace(diff) == "" {
		ui.ShowWarning("No staged changes found. Please stage your changes with 'git add' first")
		return fmt.Errorf("no staged changes found")
	}

	// Truncate diff if too long
	if maxDiffLines > 0 {
		diff = git.TruncateDiff(diff, maxDiffLines)
	}

	// Get repository context
	repoName, _ := repo.GetRepoName(ctx)
	branch, _ := repo.GetCurrentBranch(ctx)

	// Show context info if verbose
	contextFormatter := ui.NewContextFormatter()
	if info := contextFormatter.FormatRepoInfo(repoName, branch, verbose); info != "" {
		fmt.Print(info)
	}

	if verbose {
		diffLines := len(strings.Split(diff, "\n"))
		ui.ShowInfo(fmt.Sprintf("Analyzing %d lines of changes", diffLines))
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
	spinner := ui.NewStreamingSpinner("🤖 Generating commit message")
	spinner.Start()

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
			spinner.Update()
			commitMessage.WriteString(resp.Message.Content)

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
		ui.ShowError("Failed to generate commit message: " + streamErr.Error())
		return streamErr
	}

	// Clean up the generated message
	message := prompt.SanitizeCommitMessage(commitMessage.String())

	if message == "" {
		ui.ShowError("Generated commit message is empty")
		return fmt.Errorf("generated commit message is empty")
	}

	// Validate the message
	if err := prompt.ValidateCommitMessage(message); err != nil {
		ui.ShowWarning("Validation warning: " + err.Error())
	}

	// Display the generated message beautifully
	formatter := ui.NewCommitMessageFormatter()
	fmt.Print(formatter.FormatGenerated(message))

	if dryRun {
		ui.ShowInfo("Dry run mode - not committing")
		return nil
	}

	// Ask for confirmation unless auto-commit is enabled
	if !autoCommit {
		fmt.Print(formatter.FormatConfirmation())
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			ui.ShowError("Failed to read user input: " + err.Error())
			return err
		}

		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			ui.ShowInfo("Commit cancelled")
			return nil
		}
	}

	// Commit the changes
	if verbose {
		ui.ShowInfo("Committing changes...")
	}

	commitCmd := fmt.Sprintf(`git commit -m %q`, message)
	if err := runShellCommand(ctx, commitCmd); err != nil {
		ui.ShowError("Failed to commit: " + err.Error())
		return err
	}

	ui.ShowSuccess("Changes committed successfully!")
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
