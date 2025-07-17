package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient("http://localhost:11434")
	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.baseURL != "http://localhost:11434" {
		t.Errorf("Expected baseURL to be 'http://localhost:11434', got '%s'", client.baseURL)
	}
}

func TestNewClientTrimsSlash(t *testing.T) {
	client := NewClient("http://localhost:11434/")
	if client.baseURL != "http://localhost:11434" {
		t.Errorf("Expected baseURL to be 'http://localhost:11434', got '%s'", client.baseURL)
	}
}

func TestPing(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tags" {
			t.Errorf("Expected path '/api/tags', got '%s'", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"models":[]}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	ctx := context.Background()

	err := client.Ping(ctx)
	if err != nil {
		t.Errorf("Ping failed: %v", err)
	}
}

func TestPingFailure(t *testing.T) {
	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	ctx := context.Background()

	err := client.Ping(ctx)
	if err == nil {
		t.Error("Expected Ping to fail, but it succeeded")
	}
}

func TestChat(t *testing.T) {
	// Create mock server that returns streaming responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Errorf("Expected path '/api/chat', got '%s'", r.URL.Path)
		}

		if r.Method != "POST" {
			t.Errorf("Expected POST method, got '%s'", r.Method)
		}

		// Write streaming responses
		responses := []ChatResponse{
			{
				Message: Message{Content: "Hello"},
				Done:    false,
			},
			{
				Message: Message{Content: " world"},
				Done:    false,
			},
			{
				Message: Message{Content: "!"},
				Done:    true,
			},
		}

		for _, resp := range responses {
			jsonData, _ := json.Marshal(resp)
			w.Write(jsonData)
			w.Write([]byte("\n"))
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	ctx := context.Background()

	req := ChatRequest{
		Model: "test-model",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}

	respChan, errChan := client.Chat(ctx, req)

	var responses []ChatResponse
	var streamErr error

	for {
		select {
		case resp, ok := <-respChan:
			if !ok {
				goto Complete
			}
			responses = append(responses, resp)

		case err := <-errChan:
			streamErr = err
			goto Complete

		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out")
		}
	}

Complete:
	if streamErr != nil {
		t.Errorf("Chat failed: %v", streamErr)
	}

	if len(responses) != 3 {
		t.Errorf("Expected 3 responses, got %d", len(responses))
	}

	// Check that we got the expected content
	var fullMessage strings.Builder
	for _, resp := range responses {
		fullMessage.WriteString(resp.Message.Content)
	}

	expected := "Hello world!"
	if fullMessage.String() != expected {
		t.Errorf("Expected message '%s', got '%s'", expected, fullMessage.String())
	}

	// Check that the last response is marked as done
	if !responses[len(responses)-1].Done {
		t.Error("Expected last response to be marked as done")
	}
} 