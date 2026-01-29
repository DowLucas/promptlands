package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/lucas/promptlands/internal/game"
)

// Note: The game.LLMClient interface is defined in the game package.
// This file implements that interface.

// GeminiClient implements the Client interface for Google Gemini
type GeminiClient struct {
	apiKey     string
	model      string
	timeout    time.Duration
	httpClient *http.Client
	baseURL    string
}

// NewGeminiClient creates a new Gemini API client
func NewGeminiClient(apiKey, model string, timeout time.Duration) *GeminiClient {
	return &GeminiClient{
		apiKey:  apiKey,
		model:   model,
		timeout: timeout,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL: "https://generativelanguage.googleapis.com/v1beta",
	}
}

// GeminiRequest represents the request format for Gemini API
type GeminiRequest struct {
	Contents         []GeminiContent        `json:"contents"`
	GenerationConfig GeminiGenerationConfig `json:"generationConfig,omitempty"`
}

// GeminiContent represents a content block in the request
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

// GeminiPart represents a part of content
type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiGenerationConfig configures generation parameters
type GeminiGenerationConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
	ResponseMimeType string `json:"responseMimeType,omitempty"`
}

// GeminiResponse represents the response from Gemini API
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error,omitempty"`
}

// GetAction sends a prompt to Gemini and returns the parsed action
func (c *GeminiClient) GetAction(ctx context.Context, agentID uuid.UUID, prompt string) (game.Action, error) {
	if c.apiKey == "" {
		return game.WaitAction(agentID), fmt.Errorf("no API key configured")
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", c.baseURL, c.model, c.apiKey)

	reqBody := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: prompt},
				},
			},
		},
		GenerationConfig: GeminiGenerationConfig{
			Temperature:      0.7,
			MaxOutputTokens:  256,
			ResponseMimeType: "application/json",
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return game.WaitAction(agentID), fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return game.WaitAction(agentID), fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.Error("LLM request failed", "agent_id", agentID, "model", c.model, "error", err)
		return game.WaitAction(agentID), fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("LLM response read failed", "agent_id", agentID, "model", c.model, "error", err)
		return game.WaitAction(agentID), fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := resp.Header.Get("Retry-After")
		slog.Warn("LLM rate limited",
			"agent_id", agentID,
			"model", c.model,
			"status", resp.StatusCode,
			"retry_after", retryAfter,
			"body", string(body),
		)
		return game.WaitAction(agentID), fmt.Errorf("rate limited (status 429, retry-after: %s): %s", retryAfter, string(body))
	}

	if resp.StatusCode != http.StatusOK {
		slog.Error("LLM API error",
			"agent_id", agentID,
			"model", c.model,
			"status", resp.StatusCode,
			"body", string(body),
		)
		return game.WaitAction(agentID), fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		slog.Error("LLM response parse failed", "agent_id", agentID, "model", c.model, "error", err)
		return game.WaitAction(agentID), fmt.Errorf("failed to parse response: %w", err)
	}

	if geminiResp.Error != nil {
		slog.Error("LLM API returned error",
			"agent_id", agentID,
			"model", c.model,
			"error_code", geminiResp.Error.Code,
			"error_message", geminiResp.Error.Message,
		)
		return game.WaitAction(agentID), fmt.Errorf("API error: %s", geminiResp.Error.Message)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return game.WaitAction(agentID), fmt.Errorf("empty response from API")
	}

	responseText := geminiResp.Candidates[0].Content.Parts[0].Text
	return game.ParseAction(agentID, []byte(responseText))
}

// MockClient is a mock LLM client for testing
type MockClient struct{}

// NewMockClient creates a new mock client
func NewMockClient() *MockClient {
	return &MockClient{}
}

// GetAction returns a random valid action for testing
func (c *MockClient) GetAction(ctx context.Context, agentID uuid.UUID, prompt string) (game.Action, error) {
	// Simple mock logic: randomly move or claim
	actions := []game.Action{
		game.MoveAction(agentID, game.DirNorth),
		game.MoveAction(agentID, game.DirSouth),
		game.MoveAction(agentID, game.DirEast),
		game.MoveAction(agentID, game.DirWest),
		game.ClaimAction(agentID),
		game.WaitAction(agentID),
	}

	// Use time-based pseudo-random selection
	idx := int(time.Now().UnixNano()) % len(actions)
	action := actions[idx]
	action.ReceivedAt = time.Now()
	return action, nil
}
