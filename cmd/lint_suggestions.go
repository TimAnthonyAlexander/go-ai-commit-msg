package cmd

import (
	"context"
	"fmt"
	"gh-smart-commit/pkg/git"
	"gh-smart-commit/pkg/ollama"
	"gh-smart-commit/pkg/prompt"
	"gh-smart-commit/pkg/ui"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// lintSuggestionsCmd represents the lint-suggestions command
var lintSuggestionsCmd = &cobra.Command{
	Use:   "lint-suggestions",
	Short: "Get AI-powered code improvement suggestions",
	Long: `Analyze your staged or unstaged changes and provide ordered suggestions
for code improvements, best practices, and potential issues.

The suggestions are color-coded by severity and focus on:
- Code quality and maintainability
- Performance improvements
- Security considerations
- Best practice adherence`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLintSuggestions(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(lintSuggestionsCmd)

	// Command-specific flags
	lintSuggestionsCmd.Flags().Bool("staged", true, "Analyze staged changes (default)")
	lintSuggestionsCmd.Flags().Bool("unstaged", false, "Analyze unstaged changes")
	lintSuggestionsCmd.Flags().String("severity", "all", "Filter by severity: all, high, medium, low")
	lintSuggestionsCmd.Flags().Int("max-suggestions", 10, "Maximum number of suggestions to display")
}

func runLintSuggestions(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Get flags
	analyzeStaged, _ := cmd.Flags().GetBool("staged")
	analyzeUnstaged, _ := cmd.Flags().GetBool("unstaged")
	severityFilter, _ := cmd.Flags().GetString("severity")
	maxSuggestions, _ := cmd.Flags().GetInt("max-suggestions")
	verbose := viper.GetBool("verbose")

	// Validate flags
	if !analyzeStaged && !analyzeUnstaged {
		analyzeStaged = true // Default to staged if neither specified
	}

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

	// Get appropriate diff
	var diff string
	var diffType string

	if analyzeStaged {
		diff, err = repo.GetStagedDiff(ctx)
		if err != nil {
			ui.ShowError("Failed to get staged diff: " + err.Error())
			return err
		}
		diffType = "staged"
	} else {
		diff, err = repo.GetUnstagedDiff(ctx)
		if err != nil {
			ui.ShowError("Failed to get unstaged diff: " + err.Error())
			return err
		}
		diffType = "unstaged"
	}

	if strings.TrimSpace(diff) == "" {
		if analyzeStaged {
			ui.ShowWarning("No staged changes found. Please stage your changes with 'git add' first")
			return fmt.Errorf("no staged changes found")
		} else {
			ui.ShowWarning("No unstaged changes found. Please make some changes first")
			return fmt.Errorf("no unstaged changes found")
		}
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
		ui.ShowInfo(fmt.Sprintf("Analyzing %s changes (%d lines)", diffType, diffLines))
		ui.ShowInfo(fmt.Sprintf("Severity filter: %s", severityFilter))
	}

	// Build prompt
	builder := prompt.NewBuilder()
	promptCtx := prompt.Context{
		Repo:   repoName,
		Branch: branch,
		Diff:   diff,
	}

	systemPrompt, userPrompt, err := builder.Build("lint-suggestions", promptCtx)
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
	spinner := ui.NewStreamingSpinner(fmt.Sprintf("🔍 Analyzing %s changes for improvements", diffType))
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
		ui.ShowError("Failed to generate suggestions: " + streamErr.Error())
		return streamErr
	}

	response := strings.TrimSpace(responseBuilder.String())
	if response == "" {
		ui.ShowWarning("No suggestions generated")
		return fmt.Errorf("no suggestions generated")
	}

	// Parse suggestions
	suggestions := parseSuggestions(response)

	// Filter by severity
	filteredSuggestions := filterSuggestionsBySeverity(suggestions, severityFilter)

	// Limit suggestions
	if len(filteredSuggestions) > maxSuggestions {
		filteredSuggestions = filteredSuggestions[:maxSuggestions]
	}

	// Display suggestions beautifully
	formatter := ui.NewSuggestionFormatter()

	// Convert to UI suggestions format
	uiSuggestions := make([]ui.Suggestion, len(filteredSuggestions))
	for i, s := range filteredSuggestions {
		uiSuggestions[i] = ui.Suggestion{
			Severity:    s.Severity,
			Title:       s.Title,
			Description: s.Description,
			Number:      s.Number,
		}
	}

	output := formatter.FormatSuggestionsList(uiSuggestions, diffType, len(suggestions))
	fmt.Print(output)

	// Show additional info about filtering
	if severityFilter != "all" {
		ui.ShowInfo(fmt.Sprintf("Showing only %s severity suggestions", strings.ToUpper(severityFilter)))
	}

	return nil
}

// Suggestion represents a code improvement suggestion
type Suggestion struct {
	Severity    string
	Title       string
	Description string
	Number      int
}

// parseSuggestions parses the AI response into structured suggestions
func parseSuggestions(response string) []Suggestion {
	var suggestions []Suggestion

	// Split response into blocks and parse each numbered item
	lines := strings.Split(response, "\n")
	var currentSuggestion *Suggestion

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Pattern to match numbered suggestions with severity: "1. [HIGH] Title"
		pattern := regexp.MustCompile(`^(\d+)\.\s*\[([^\]]+)\]\s*(.+)`)
		if matches := pattern.FindStringSubmatch(line); matches != nil {
			// Save previous suggestion if exists
			if currentSuggestion != nil {
				suggestions = append(suggestions, *currentSuggestion)
			}

			// Start new suggestion
			number, _ := strconv.Atoi(matches[1])
			severity := strings.TrimSpace(strings.ToUpper(matches[2]))
			title := strings.TrimSpace(matches[3])

			currentSuggestion = &Suggestion{
				Number:   number,
				Severity: severity,
				Title:    title,
			}
		} else if currentSuggestion != nil && line != "" {
			// Add to description of current suggestion
			if currentSuggestion.Description == "" {
				currentSuggestion.Description = line
			} else {
				currentSuggestion.Description += " " + line
			}
		}
	}

	// Don't forget the last suggestion
	if currentSuggestion != nil {
		suggestions = append(suggestions, *currentSuggestion)
	}

	// Fallback: simple line-by-line parsing if regex doesn't work
	if len(suggestions) == 0 {
		lines := strings.Split(response, "\n")
		number := 1

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// Look for severity indicators
			severity := "MEDIUM"
			if strings.Contains(strings.ToUpper(line), "[HIGH]") {
				severity = "HIGH"
			} else if strings.Contains(strings.ToUpper(line), "[LOW]") {
				severity = "LOW"
			}

			// Clean up the line
			line = regexp.MustCompile(`^\d+\.\s*`).ReplaceAllString(line, "")
			line = regexp.MustCompile(`\[(?:HIGH|MEDIUM|LOW)\]\s*`).ReplaceAllString(line, "")

			if line != "" {
				suggestions = append(suggestions, Suggestion{
					Number:   number,
					Severity: severity,
					Title:    line,
				})
				number++
			}
		}
	}

	return suggestions
}

// filterSuggestionsBySeverity filters suggestions by severity level
func filterSuggestionsBySeverity(suggestions []Suggestion, severityFilter string) []Suggestion {
	if severityFilter == "all" {
		return suggestions
	}

	var filtered []Suggestion
	targetSeverity := strings.ToUpper(severityFilter)

	for _, suggestion := range suggestions {
		if suggestion.Severity == targetSeverity {
			filtered = append(filtered, suggestion)
		}
	}

	return filtered
}
