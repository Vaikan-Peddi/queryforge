package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"queryforge/backend/internal/models"
)

type AIClient struct {
	baseURL string
	client  *http.Client
}

type GenerateRequest struct {
	Question    string                `json:"question"`
	Schema      models.DatabaseSchema `json:"schema"`
	SafetyRules []string              `json:"safety_rules"`
}

type GenerateResponse struct {
	SQL         string  `json:"sql"`
	Explanation string  `json:"explanation"`
	Confidence  float64 `json:"confidence"`
}

type AIHealth struct {
	Status   string `json:"status"`
	Provider string `json:"provider"`
	Model    string `json:"model"`
	BaseURL  string `json:"base_url"`
}

func NewAIClient(baseURL string, timeout time.Duration) *AIClient {
	if timeout <= 0 {
		timeout = 130 * time.Second
	}
	return &AIClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: timeout},
	}
}

func (c *AIClient) GenerateSQL(ctx context.Context, req GenerateRequest) (GenerateResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return GenerateResponse{}, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/generate-sql", bytes.NewReader(body))
	if err != nil {
		return GenerateResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return GenerateResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return GenerateResponse{}, fmt.Errorf("ai service returned status %d", resp.StatusCode)
	}
	var out GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return GenerateResponse{}, err
	}
	return out, nil
}

func (c *AIClient) Health(ctx context.Context) (AIHealth, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return AIHealth{}, err
	}
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return AIHealth{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return AIHealth{}, fmt.Errorf("ai service returned status %d", resp.StatusCode)
	}
	var out AIHealth
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return AIHealth{}, err
	}
	return out, nil
}
