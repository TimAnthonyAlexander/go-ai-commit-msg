package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Color palette - terminal-adaptive colors that work on both light and dark backgrounds
	primaryColor = lipgloss.Color("#007AFF") // Blue
	successColor = lipgloss.Color("#34C759") // Green
	warningColor = lipgloss.Color("#FF9500") // Orange
	errorColor   = lipgloss.Color("#FF3B30") // Red
	mutedColor   = lipgloss.Color("#8E8E93") // Gray
	accentColor  = lipgloss.Color("#5856D6") // Purple

	// Use adaptive colors that work on any background
	adaptiveTextColor = lipgloss.AdaptiveColor{
		Light: "#1C1C1E", // Dark text on light background
		Dark:  "#FFFFFF", // White text on dark background
	}
	adaptiveMutedColor = lipgloss.AdaptiveColor{
		Light: "#8E8E93", // Gray on light background
		Dark:  "#98989D", // Lighter gray on dark background
	}
	adaptiveBackgroundColor = lipgloss.AdaptiveColor{
		Light: "#FFFFFF", // White on light terminals
		Dark:  "#1C1C1E", // Dark on dark terminals
	}
	adaptiveBorderColor = lipgloss.AdaptiveColor{
		Light: "#E5E5EA", // Light gray border on light background
		Dark:  "#38383A", // Dark gray border on dark background
	}
)

// Base styles
var (
	// Typography
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(adaptiveTextColor).
			PaddingTop(1).
			PaddingBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(adaptiveMutedColor).
			PaddingBottom(1)

	BodyStyle = lipgloss.NewStyle().
			Foreground(adaptiveTextColor)

	// Layout containers
	ContainerStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(adaptiveBorderColor).
			Foreground(adaptiveTextColor)

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			PaddingTop(1).
			PaddingBottom(1).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(adaptiveBorderColor)

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
			Foreground(adaptiveMutedColor)

	// Interactive elements
	ButtonStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Background(primaryColor).
			Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"}).
			Bold(true).
			Border(lipgloss.RoundedBorder())

	SecondaryButtonStyle = lipgloss.NewStyle().
				Padding(0, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor).
				Foreground(primaryColor)

	// Code and data
	CodeStyle = lipgloss.NewStyle().
			Background(lipgloss.AdaptiveColor{Light: "#F2F2F7", Dark: "#2C2C2E"}).
			Foreground(adaptiveTextColor).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(adaptiveBorderColor)

	CommitMessageStyle = lipgloss.NewStyle().
				Foreground(adaptiveTextColor).
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
		Foreground(adaptiveTextColor)

	content := fmt.Sprintf("%s %s",
		SuccessStyle.Render("✓"),
		BodyStyle.Render(message))

	return style.Render(content)
}

// RenderErrorBox renders an error message in a red box
func RenderErrorBox(message string) string {
	style := ContainerStyle.Copy().
		BorderForeground(errorColor).
		Foreground(adaptiveTextColor)

	content := fmt.Sprintf("%s %s",
		ErrorStyle.Render("✗"),
		BodyStyle.Render(message))

	return style.Render(content)
}

// RenderWarningBox renders a warning message in an orange box
func RenderWarningBox(message string) string {
	style := ContainerStyle.Copy().
		BorderForeground(warningColor).
		Foreground(adaptiveTextColor)

	content := fmt.Sprintf("%s %s",
		WarningStyle.Render("⚠"),
		BodyStyle.Render(message))

	return style.Render(content)
}
