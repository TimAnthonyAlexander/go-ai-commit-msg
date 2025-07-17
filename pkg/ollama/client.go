package ollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client represents an Ollama HTTP client
type Client struct {
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
}

// ChatRequest represents a chat request to Ollama
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
	Options  Options   `json:"options,omitempty"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Options represents model options
type Options struct {
	Temperature float32 `json:"temperature,omitempty"`
}

// ChatResponse represents a streaming chat response
type ChatResponse struct {
	Model              string    `json:"model"`
	CreatedAt          time.Time `json:"created_at"`
	Message            Message   `json:"message"`
	Done               bool      `json:"done"`
	TotalDuration      int64     `json:"total_duration,omitempty"`
	LoadDuration       int64     `json:"load_duration,omitempty"`
	PromptEvalCount    int       `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64     `json:"prompt_eval_duration,omitempty"`
	EvalCount          int       `json:"eval_count,omitempty"`
	EvalDuration       int64     `json:"eval_duration,omitempty"`
}

// NewClient creates a new Ollama client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     90 * time.Second,
				DisableCompression:  false,
				DisableKeepAlives:   false,
			},
		},
		timeout: 5 * time.Minute, // Longer timeout for LLM responses
	}
}

// Chat sends a chat request and returns a channel for streaming responses
func (c *Client) Chat(ctx context.Context, req ChatRequest) (<-chan ChatResponse, <-chan error) {
	respChan := make(chan ChatResponse, 10)
	errChan := make(chan error, 1)

	go func() {
		defer close(respChan)
		defer close(errChan)

		if err := c.streamChat(ctx, req, respChan); err != nil {
			errChan <- err
		}
	}()

	return respChan, errChan
}

// streamChat performs the actual streaming request
func (c *Client) streamChat(ctx context.Context, req ChatRequest, respChan chan<- ChatResponse) error {
	// Create request context with timeout
	reqCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Ensure streaming is enabled
	req.Stream = true

	// Marshal request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(reqCtx, "POST", c.baseURL+"/api/chat", bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// Execute request with retry logic
	resp, err := c.executeWithRetry(httpReq, 3)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ollama request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Stream responses
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var chatResp ChatResponse
		if err := json.Unmarshal([]byte(line), &chatResp); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}

		select {
		case respChan <- chatResp:
		case <-ctx.Done():
			return ctx.Err()
		}

		// Stop if done
		if chatResp.Done {
			break
		}
	}

	return scanner.Err()
}

// executeWithRetry executes an HTTP request with exponential backoff retry
func (c *Client) executeWithRetry(req *http.Request, maxRetries int) (*http.Response, error) {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		resp, err := c.httpClient.Do(req)
		if err == nil && resp.StatusCode < 500 {
			return resp, nil
		}

		if resp != nil {
			resp.Body.Close()
		}
		lastErr = err

		// Don't retry on the last attempt
		if i == maxRetries-1 {
			break
		}

		// Exponential backoff: 1s, 2s, 4s
		backoff := time.Duration(1<<uint(i)) * time.Second
		time.Sleep(backoff)
	}

	return nil, lastErr
}

// Ping checks if the Ollama server is accessible
func (c *Client) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("failed to create ping request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to ping ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama ping failed with status %d", resp.StatusCode)
	}

	return nil
} 