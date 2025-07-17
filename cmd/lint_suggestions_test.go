package cmd

import (
	"strings"
	"testing"
)

func TestParseSuggestions(t *testing.T) {
	// Test response in the expected format
	response := `1. [HIGH] Add error handling for division by zero
   The divide function should check if the divisor is zero before performing division.

2. [MEDIUM] Use range loop instead of traditional for loop
   Replace the traditional for loop with a range loop for better readability.

3. [LOW] Add documentation for public functions
   Functions should have proper Go documentation comments.`

	suggestions := parseSuggestions(response)

	if len(suggestions) != 3 {
		t.Errorf("Expected 3 suggestions, got %d", len(suggestions))
	}

	// Test first suggestion
	if suggestions[0].Severity != "HIGH" {
		t.Errorf("Expected first suggestion severity 'HIGH', got '%s'", suggestions[0].Severity)
	}

	if !strings.Contains(suggestions[0].Title, "Add error handling") {
		t.Errorf("Expected first suggestion title to contain 'Add error handling', got '%s'", suggestions[0].Title)
	}

	if !strings.Contains(suggestions[0].Description, "divide function") {
		t.Errorf("Expected first suggestion description to contain 'divide function', got '%s'", suggestions[0].Description)
	}

	// Test second suggestion
	if suggestions[1].Severity != "MEDIUM" {
		t.Errorf("Expected second suggestion severity 'MEDIUM', got '%s'", suggestions[1].Severity)
	}

	// Test third suggestion
	if suggestions[2].Severity != "LOW" {
		t.Errorf("Expected third suggestion severity 'LOW', got '%s'", suggestions[2].Severity)
	}
}

func TestParseSuggestionsSimpleFormat(t *testing.T) {
	// Test fallback parsing with simple format
	response := `[HIGH] Critical security issue found
[MEDIUM] Performance can be improved
[LOW] Code style suggestion
Additional context line`

	suggestions := parseSuggestions(response)

	if len(suggestions) < 3 {
		t.Errorf("Expected at least 3 suggestions, got %d", len(suggestions))
	}

	// Check that severities are detected
	foundHigh := false
	foundMedium := false
	foundLow := false

	for _, suggestion := range suggestions {
		switch suggestion.Severity {
		case "HIGH":
			foundHigh = true
		case "MEDIUM":
			foundMedium = true
		case "LOW":
			foundLow = true
		}
	}

	if !foundHigh {
		t.Error("Expected to find HIGH severity suggestion")
	}
	if !foundMedium {
		t.Error("Expected to find MEDIUM severity suggestion")
	}
	if !foundLow {
		t.Error("Expected to find LOW severity suggestion")
	}
}

func TestFilterSuggestionsBySeverity(t *testing.T) {
	suggestions := []Suggestion{
		{Severity: "HIGH", Title: "Critical issue"},
		{Severity: "MEDIUM", Title: "Moderate issue"},
		{Severity: "LOW", Title: "Minor issue"},
		{Severity: "HIGH", Title: "Another critical issue"},
	}

	// Test filtering by HIGH
	highOnly := filterSuggestionsBySeverity(suggestions, "high")
	if len(highOnly) != 2 {
		t.Errorf("Expected 2 HIGH suggestions, got %d", len(highOnly))
	}

	// Test filtering by MEDIUM
	mediumOnly := filterSuggestionsBySeverity(suggestions, "medium")
	if len(mediumOnly) != 1 {
		t.Errorf("Expected 1 MEDIUM suggestion, got %d", len(mediumOnly))
	}

	// Test filtering by all
	allSuggestions := filterSuggestionsBySeverity(suggestions, "all")
	if len(allSuggestions) != 4 {
		t.Errorf("Expected 4 suggestions for 'all', got %d", len(allSuggestions))
	}

	// Test case insensitivity
	highOnlyLower := filterSuggestionsBySeverity(suggestions, "HIGH")
	if len(highOnlyLower) != 2 {
		t.Errorf("Expected 2 HIGH suggestions (case insensitive), got %d", len(highOnlyLower))
	}
}

func TestParseSuggestionsEmpty(t *testing.T) {
	suggestions := parseSuggestions("")
	if len(suggestions) != 0 {
		t.Errorf("Expected 0 suggestions for empty input, got %d", len(suggestions))
	}
}

func TestParseSuggestionsNoMatches(t *testing.T) {
	response := "This is just some random text without any structured suggestions."
	suggestions := parseSuggestions(response)

	// Should fall back to simple parsing and create at least one suggestion
	if len(suggestions) == 0 {
		t.Error("Expected fallback parsing to create at least one suggestion")
	}
}
