package claude

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"realmweaver/engine"
)

// Client handles communication with the Claude API.
type Client struct {
	apiURL     string
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewClient creates a Claude API client from environment variables.
func NewClient() *Client {
	apiURL := os.Getenv("CLAUDE_API_URL")
	if apiURL == "" {
		apiURL = "https://ai-proxy.9635783.xyz/v1/messages"
	}
	apiKey := os.Getenv("CLAUDE_API_KEY")
	if apiKey == "" {
		panic("CLAUDE_API_KEY environment variable is required")
	}
	model := os.Getenv("CLAUDE_MODEL")
	if model == "" {
		model = "claude-sonnet-4-6"
	}

	return &Client{
		apiURL: apiURL,
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Request represents a Claude API request.
type Request struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system,omitempty"`
	Messages  []Message `json:"messages"`
	Stream    bool      `json:"stream,omitempty"`
}

// Message is a conversation turn.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// apiResponse is the non-streaming API response.
type apiResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

// streamEvent is a single SSE event from the streaming API.
type streamEvent struct {
	Type  string          `json:"type"`
	Delta json.RawMessage `json:"delta,omitempty"`
	Index int             `json:"index,omitempty"`
}

type contentDelta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Generate sends a non-streaming request and returns the parsed response.
func (c *Client) Generate(systemPrompt, userMessage string) (*engine.ClaudeResponse, error) {
	req := Request{
		Model:     c.model,
		MaxTokens: 1024,
		System:    systemPrompt,
		Messages: []Message{
			{Role: "user", Content: userMessage},
		},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	var apiResp apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(apiResp.Content) == 0 {
		return nil, fmt.Errorf("empty response from Claude")
	}

	return parseClaudeResponse(apiResp.Content[0].Text), nil
}

// GenerateStream sends a streaming request and calls onToken for each text chunk.
// Returns the full assembled response when done.
func (c *Client) GenerateStream(systemPrompt, userMessage string, onToken func(token string)) (*engine.ClaudeResponse, error) {
	req := Request{
		Model:     c.model,
		MaxTokens: 1024,
		System:    systemPrompt,
		Messages: []Message{
			{Role: "user", Content: userMessage},
		},
		Stream: true,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	var fullText strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var event streamEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		if event.Type == "content_block_delta" {
			var delta contentDelta
			if err := json.Unmarshal(event.Delta, &delta); err == nil && delta.Text != "" {
				fullText.WriteString(delta.Text)
				if onToken != nil {
					onToken(delta.Text)
				}
			}
		}
	}

	return parseClaudeResponse(fullText.String()), nil
}

// lenientResponse is used for parsing when Claude's hints format doesn't match our struct.
type lenientResponse struct {
	Narrative string          `json:"narrative"`
	Hints     json.RawMessage `json:"hints"`
}

// parseClaudeResponse tries to extract structured JSON from Claude's response.
// Falls back to treating the whole response as narrative.
func parseClaudeResponse(text string) *engine.ClaudeResponse {
	// Strip markdown code blocks if present
	cleaned := text
	if idx := strings.Index(cleaned, "```json"); idx >= 0 {
		cleaned = cleaned[idx+7:]
	} else if idx := strings.Index(cleaned, "```"); idx >= 0 {
		cleaned = cleaned[idx+3:]
	}
	if idx := strings.LastIndex(cleaned, "```"); idx >= 0 {
		cleaned = cleaned[:idx]
	}
	cleaned = strings.TrimSpace(cleaned)

	// Find JSON object
	jsonStr := cleaned
	start := strings.Index(cleaned, "{")
	end := strings.LastIndex(cleaned, "}")
	if start >= 0 && end > start {
		jsonStr = cleaned[start : end+1]
	}

	// Parse with lenient struct (accepts any hints format)
	var lenient lenientResponse
	if err := json.Unmarshal([]byte(jsonStr), &lenient); err == nil && lenient.Narrative != "" {
		resp := &engine.ClaudeResponse{
			Narrative: lenient.Narrative,
			Hints:     engine.ClaudeHints{},
		}
		// Try to parse hints into our struct (ignore errors)
		if lenient.Hints != nil {
			json.Unmarshal(lenient.Hints, &resp.Hints)
		}
		return resp
	}

	// Fallback: treat entire original text as narrative, strip JSON/markdown
	fallback := text
	fallback = strings.ReplaceAll(fallback, "```json", "")
	fallback = strings.ReplaceAll(fallback, "```", "")
	fallback = strings.TrimSpace(fallback)
	return &engine.ClaudeResponse{
		Narrative: fallback,
		Hints:     engine.ClaudeHints{},
	}
}
