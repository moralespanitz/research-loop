// Package llm provides a model-agnostic LLM client for Research Loop.
// Supports Anthropic, OpenAI-compatible endpoints, and Ollama.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/research-loop/research-loop/internal/config"
)

// Message is a single chat turn.
type Message struct {
	Role    string `json:"role"`    // "user" | "assistant" | "system"
	Content string `json:"content"`
}

// Client is the interface every LLM backend implements.
type Client interface {
	Complete(ctx context.Context, system string, messages []Message) (string, error)
	ModelName() string
}

// New returns the appropriate Client based on the config.
func New(cfg config.LLMConfig) (Client, error) {
	switch strings.ToLower(cfg.Provider) {
	case "anthropic":
		return newAnthropic(cfg)
	case "openai":
		return newOpenAI(cfg)
	case "ollama":
		return newOllama(cfg)
	default:
		return newOpenAI(cfg) // treat unknown as openai-compatible
	}
}

// ─── Anthropic ───────────────────────────────────────────────────────────────

type anthropicClient struct {
	apiKey string
	model  string
	http   *http.Client
}

func newAnthropic(cfg config.LLMConfig) (*anthropicClient, error) {
	key := os.Getenv(cfg.APIKeyEnv)
	if key == "" {
		key = os.Getenv("ANTHROPIC_API_KEY")
	}
	if key == "" {
		return nil, fmt.Errorf("Anthropic API key not found in env var %q", cfg.APIKeyEnv)
	}
	model := cfg.Model
	if model == "" {
		model = "claude-sonnet-4-5"
	}
	return &anthropicClient{
		apiKey: key,
		model:  model,
		http:   &http.Client{Timeout: 120 * time.Second},
	}, nil
}

func (c *anthropicClient) ModelName() string { return c.model }

func (c *anthropicClient) Complete(ctx context.Context, system string, messages []Message) (string, error) {
	type anthropicMsg struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type requestBody struct {
		Model     string         `json:"model"`
		MaxTokens int            `json:"max_tokens"`
		System    string         `json:"system,omitempty"`
		Messages  []anthropicMsg `json:"messages"`
	}

	msgs := make([]anthropicMsg, len(messages))
	for i, m := range messages {
		msgs[i] = anthropicMsg{Role: m.Role, Content: m.Content}
	}

	body := requestBody{
		Model:     c.model,
		MaxTokens: 4096,
		System:    system,
		Messages:  msgs,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("anthropic request: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("anthropic %d: %s", resp.StatusCode, string(raw))
	}

	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", fmt.Errorf("parsing anthropic response: %w", err)
	}
	if len(result.Content) == 0 {
		return "", fmt.Errorf("empty response from anthropic")
	}
	return result.Content[0].Text, nil
}

// ─── OpenAI-compatible ───────────────────────────────────────────────────────

type openAIClient struct {
	apiKey  string
	model   string
	baseURL string
	http    *http.Client
}

func newOpenAI(cfg config.LLMConfig) (*openAIClient, error) {
	key := os.Getenv(cfg.APIKeyEnv)
	if key == "" {
		key = os.Getenv("OPENAI_API_KEY")
	}
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	model := cfg.Model
	if model == "" {
		model = "gpt-4o"
	}
	return &openAIClient{
		apiKey:  key,
		model:   model,
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: 120 * time.Second},
	}, nil
}

func (c *openAIClient) ModelName() string { return c.model }

func (c *openAIClient) Complete(ctx context.Context, system string, messages []Message) (string, error) {
	type oaiMsg struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type requestBody struct {
		Model    string   `json:"model"`
		Messages []oaiMsg `json:"messages"`
	}

	var msgs []oaiMsg
	if system != "" {
		msgs = append(msgs, oaiMsg{Role: "system", Content: system})
	}
	for _, m := range messages {
		msgs = append(msgs, oaiMsg{Role: m.Role, Content: m.Content})
	}

	body := requestBody{Model: c.model, Messages: msgs}
	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("openai request: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("openai %d: %s", resp.StatusCode, string(raw))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", fmt.Errorf("parsing openai response: %w", err)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("empty response from openai")
	}
	return result.Choices[0].Message.Content, nil
}

// ─── Ollama ──────────────────────────────────────────────────────────────────

type ollamaClient struct {
	model   string
	baseURL string
	http    *http.Client
}

func newOllama(cfg config.LLMConfig) (*ollamaClient, error) {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	model := cfg.Model
	if model == "" {
		model = "llama3"
	}
	return &ollamaClient{
		model:   model,
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: 300 * time.Second},
	}, nil
}

func (c *ollamaClient) ModelName() string { return c.model }

func (c *ollamaClient) Complete(ctx context.Context, system string, messages []Message) (string, error) {
	// Build a single prompt from messages (Ollama generate API)
	var prompt strings.Builder
	if system != "" {
		prompt.WriteString("System: ")
		prompt.WriteString(system)
		prompt.WriteString("\n\n")
	}
	for _, m := range messages {
		prompt.WriteString(strings.Title(m.Role))
		prompt.WriteString(": ")
		prompt.WriteString(m.Content)
		prompt.WriteString("\n")
	}
	prompt.WriteString("Assistant: ")

	type requestBody struct {
		Model  string `json:"model"`
		Prompt string `json:"prompt"`
		Stream bool   `json:"stream"`
	}
	body := requestBody{Model: c.model, Prompt: prompt.String(), Stream: false}
	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama request: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("ollama %d: %s", resp.StatusCode, string(raw))
	}

	var result struct {
		Response string `json:"response"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", fmt.Errorf("parsing ollama response: %w", err)
	}
	return result.Response, nil
}
