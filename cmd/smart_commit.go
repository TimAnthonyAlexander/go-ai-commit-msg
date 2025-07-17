package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
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
	fmt.Println("smart-commit command executed (placeholder)")
	// TODO: Implement in Phase 3
	return nil
} 