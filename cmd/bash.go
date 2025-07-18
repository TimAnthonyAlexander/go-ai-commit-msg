package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"gh-smart-commit/pkg/git"
	"gh-smart-commit/pkg/ollama"
	"gh-smart-commit/pkg/prompt"
	"gh-smart-commit/pkg/ui"
)

// bashCmd represents the bash command
var bashCmd = &cobra.Command{
	Use:   "bash [description]",
	Short: "Generate and execute bash commands from descriptions",
	Long: `Generate bash commands using AI based on your description and current system context.

The AI will analyze your request along with:
- Current directory and folder structure
- Git repository status (if applicable)
- Operating system and architecture
- Available tools and environment

You'll be asked to confirm before executing the generated command.

Examples:
  gh-smart-commit bash "list all Go files in this project"
  gh-smart-commit bash "find files larger than 10MB"
  gh-smart-commit bash "create a backup of the src directory"`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBash(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(bashCmd)

	// Command-specific flags
	bashCmd.Flags().Bool("dry-run", false, "Show generated command without executing")
	bashCmd.Flags().Bool("auto-execute", false, "Execute command without confirmation (dangerous!)")
}

func runBash(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Get flags
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	autoExecute, _ := cmd.Flags().GetBool("auto-execute")
	verbose := viper.GetBool("verbose")

	// Join args to form the description
	description := strings.Join(args, " ")
	if strings.TrimSpace(description) == "" {
		ui.ShowError("Please provide a description of what you want to do")
		return fmt.Errorf("description is required")
	}

	if verbose {
		ui.ShowInfo(fmt.Sprintf("Task: %s", description))
	}

	// Gather system context
	systemCtx, err := gatherSystemContext(ctx)
	if err != nil {
		ui.ShowWarning("Failed to gather full system context: " + err.Error())
		// Continue with partial context
	}

	if verbose {
		ui.ShowInfo("Gathered system context")
	}

	// Build prompt
	builder := prompt.NewBuilder()
	promptCtx := prompt.Context{
		Repo:        systemCtx.Repo,
		Branch:      systemCtx.Branch,
		Description: description,
		SystemInfo:  systemCtx,
	}

	systemPrompt, userPrompt, err := builder.Build("bash", promptCtx)
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
	spinner := ui.NewStreamingSpinner("🧠 Generating bash command")
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
		ui.ShowError("Failed to generate bash command: " + streamErr.Error())
		return streamErr
	}

	// Clean up the generated command
	command := prompt.SanitizeBashCommand(responseBuilder.String())

	if command == "" {
		ui.ShowError("Generated command is empty")
		return fmt.Errorf("generated command is empty")
	}

	// Display the generated command beautifully
	formatter := ui.NewBashCommandFormatter()
	fmt.Print(formatter.FormatGenerated(command))

	if dryRun {
		ui.ShowInfo("Dry run mode - not executing command")
		return nil
	}

	// Ask for confirmation unless auto-execute is enabled
	if !autoExecute {
		fmt.Print(formatter.FormatConfirmation())
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			ui.ShowError("Failed to read user input: " + err.Error())
			return err
		}

		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			ui.ShowInfo("Command execution cancelled")
			return nil
		}
	}

	// Execute the command
	if verbose {
		ui.ShowInfo("Executing command...")
	}

	if err := runShellCommand(ctx, command); err != nil {
		ui.ShowError("Failed to execute command: " + err.Error())
		return err
	}

	ui.ShowSuccess("Command executed successfully!")
	return nil
}

// SystemContext holds system information for command generation
type SystemContext struct {
	OS         string
	Arch       string
	WorkingDir string
	IsGitRepo  bool
	Repo       string
	Branch     string
	FileTree   string
	Shell      string
	User       string
}

// gatherSystemContext collects system information for the prompt
func gatherSystemContext(ctx context.Context) (*SystemContext, error) {
	sysCtx := &SystemContext{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}

	// Get working directory
	if wd, err := os.Getwd(); err == nil {
		sysCtx.WorkingDir = wd
	}

	// Get current user
	if user := os.Getenv("USER"); user != "" {
		sysCtx.User = user
	} else if user := os.Getenv("USERNAME"); user != "" {
		sysCtx.User = user
	}

	// Get shell
	if shell := os.Getenv("SHELL"); shell != "" {
		sysCtx.Shell = shell
	}

	// Check if we're in a git repository
	repo := git.NewLocalRepo(".")
	if isGit, err := repo.IsInsideWorkTree(ctx); err == nil && isGit {
		sysCtx.IsGitRepo = true

		if repoName, err := repo.GetRepoName(ctx); err == nil {
			sysCtx.Repo = repoName
		}

		if branch, err := repo.GetCurrentBranch(ctx); err == nil {
			sysCtx.Branch = branch
		}
	}

	// Get a basic file tree (limited depth for context)
	if fileTree, err := getFileTree(sysCtx.WorkingDir, 2); err == nil {
		sysCtx.FileTree = fileTree
	}

	return sysCtx, nil
}

// getFileTree generates a basic file tree for context
func getFileTree(dir string, maxDepth int) (string, error) {
	if maxDepth <= 0 {
		return "", nil
	}

	var result strings.Builder
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	count := 0
	for _, entry := range entries {
		// Skip hidden files and common ignore patterns
		if strings.HasPrefix(entry.Name(), ".") ||
			entry.Name() == "node_modules" ||
			entry.Name() == "vendor" ||
			entry.Name() == "__pycache__" {
			continue
		}

		if count > 10 { // Limit entries to avoid huge output
			result.WriteString("... (more files)\n")
			break
		}

		if entry.IsDir() {
			result.WriteString(fmt.Sprintf("%s/\n", entry.Name()))
		} else {
			result.WriteString(fmt.Sprintf("%s\n", entry.Name()))
		}
		count++
	}

	return result.String(), nil
}
