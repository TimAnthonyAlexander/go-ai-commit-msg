package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Color palette - minimalist, Apple-esque
	primaryColor    = lipgloss.Color("#007AFF") // Blue
	successColor    = lipgloss.Color("#34C759") // Green
	warningColor    = lipgloss.Color("#FF9500") // Orange
	errorColor      = lipgloss.Color("#FF3B30") // Red
	mutedColor      = lipgloss.Color("#8E8E93") // Gray
	backgroundColor = lipgloss.Color("#F2F2F7") // Light gray
	surfaceColor    = lipgloss.Color("#FFFFFF") // White
	textColor       = lipgloss.Color("#1C1C1E") // Dark
	accentColor     = lipgloss.Color("#5856D6") // Purple
)

// Base styles
var (
	// Typography
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(textColor).
			PaddingTop(1).
			PaddingBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			PaddingBottom(1)

	BodyStyle = lipgloss.NewStyle().
			Foreground(textColor)

	// Layout containers
	ContainerStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(mutedColor).
			Background(surfaceColor)

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			PaddingTop(1).
			PaddingBottom(1).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(mutedColor)

	// Status styles
	SuccessStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(successColor)

	ErrorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(errorColor)

	WarningStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(warningColor)

	InfoStyle = lipgloss.NewStyle().
			Foreground(primaryColor)

	MutedStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	// Interactive elements
	ButtonStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Background(primaryColor).
			Foreground(surfaceColor).
			Bold(true).
			Border(lipgloss.RoundedBorder())

	SecondaryButtonStyle = lipgloss.NewStyle().
				Padding(0, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor).
				Foreground(primaryColor)

	// Code and data
	CodeStyle = lipgloss.NewStyle().
			Background(backgroundColor).
			Foreground(textColor).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(mutedColor)

	CommitMessageStyle = lipgloss.NewStyle().
				Background(surfaceColor).
				Foreground(textColor).
				Padding(1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor).
				Bold(true)

	// Severity indicators
	HighSeverityStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(errorColor)

	MediumSeverityStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(warningColor)

	LowSeverityStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(successColor)
)

// GetSeverityStyle returns the appropriate style for a severity level
func GetSeverityStyle(severity string) lipgloss.Style {
	switch strings.ToUpper(severity) {
	case "HIGH":
		return HighSeverityStyle
	case "MEDIUM":
		return MediumSeverityStyle
	case "LOW":
		return LowSeverityStyle
	default:
		return BodyStyle
	}
}

// GetSeverityIcon returns the appropriate icon for a severity level
func GetSeverityIcon(severity string) string {
	if IsNoColor() {
		switch strings.ToUpper(severity) {
		case "HIGH":
			return "[HIGH]"
		case "MEDIUM":
			return "[MED]"
		case "LOW":
			return "[LOW]"
		default:
			return ""
		}
	}

	switch strings.ToUpper(severity) {
	case "HIGH":
		return "🔴"
	case "MEDIUM":
		return "🟡"
	case "LOW":
		return "🟢"
	default:
		return "⚪"
	}
}

// IsNoColor checks if color output should be disabled
func IsNoColor() bool {
	return os.Getenv("NO_COLOR") != ""
}

// CreateSeparator creates a styled separator line
func CreateSeparator(width int) string {
	if width <= 0 {
		width = 60
	}
	return MutedStyle.Render(strings.Repeat("─", width))
}

// CreateDivider creates a thick divider
func CreateDivider(width int) string {
	if width <= 0 {
		width = 60
	}
	return MutedStyle.Render(strings.Repeat("━", width))
}

// RenderBox renders content in a styled box
func RenderBox(title, content string) string {
	titleRendered := HeaderStyle.Render(title)
	contentRendered := BodyStyle.Render(content)

	box := ContainerStyle.Render(
		fmt.Sprintf("%s\n\n%s", titleRendered, contentRendered),
	)

	return box
}

// RenderSuccessBox renders a success message in a green box
func RenderSuccessBox(message string) string {
	style := ContainerStyle.Copy().
		BorderForeground(successColor).
		Background(surfaceColor)

	content := fmt.Sprintf("%s %s",
		SuccessStyle.Render("✓"),
		BodyStyle.Render(message))

	return style.Render(content)
}

// RenderErrorBox renders an error message in a red box
func RenderErrorBox(message string) string {
	style := ContainerStyle.Copy().
		BorderForeground(errorColor).
		Background(surfaceColor)

	content := fmt.Sprintf("%s %s",
		ErrorStyle.Render("✗"),
		BodyStyle.Render(message))

	return style.Render(content)
}

// RenderWarningBox renders a warning message in an orange box
func RenderWarningBox(message string) string {
	style := ContainerStyle.Copy().
		BorderForeground(warningColor).
		Background(surfaceColor)

	content := fmt.Sprintf("%s %s",
		WarningStyle.Render("⚠"),
		BodyStyle.Render(message))

	return style.Render(content)
}
