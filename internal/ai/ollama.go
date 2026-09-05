package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type OllamaClient struct {
	BaseURL string
	HTTP    *http.Client
}
type GenerateRequest struct {
	Model   string         `json:"model"`
	Prompt  string         `json:"prompt"`
	Stream  bool           `json:"stream"`
	Options map[string]any `json:"options,omitempty"`
}
type GenerateResponse struct {
	Response      string `json:"response"`
	Done          bool   `json:"done"`
	TotalDuration int64  `json:"total_duration,omitempty"`
}

func NewOllamaClient(baseURL string) *OllamaClient {
	return &OllamaClient{BaseURL: strings.TrimRight(baseURL, "/"), HTTP: &http.Client{Timeout: 10 * time.Minute}}
}
func (c *OllamaClient) Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error) {
	if c.BaseURL == "" || req.Model == "" {
		return GenerateResponse{}, fmt.Errorf("ollama base URL and model are required")
	}
	req.Stream = false
	b, _ := json.Marshal(req)
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/generate", bytes.NewReader(b))
	if err != nil {
		return GenerateResponse{}, err
	}
	r.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(r)
	if err != nil {
		return GenerateResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return GenerateResponse{}, fmt.Errorf("ollama returned %s", resp.Status)
	}
	var out GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return GenerateResponse{}, err
	}
	return out, nil
}
