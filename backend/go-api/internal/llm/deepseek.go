package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"stock-analysis-api/backend/go-api/config"
)

type DeepSeekClient struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

type deepSeekRequest struct {
	Model       string                   `json:"model"`
	Messages    []deepSeekMessage        `json:"messages"`
	Stream      bool                     `json:"stream"`
	Temperature float64                  `json:"temperature"`
	MaxTokens   int                      `json:"max_tokens"`
}

type deepSeekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type deepSeekStreamResponse struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

func NewDeepSeekClient() *DeepSeekClient {
	return &DeepSeekClient{
		apiKey:  config.AppConfig.DeepSeekAPIKey,
		baseURL: config.AppConfig.DeepSeekBaseURL,
		model:   config.AppConfig.DeepSeekModel,
		client:  &http.Client{},
	}
}

func (d *DeepSeekClient) StreamAnalyze(ctx context.Context, step AnalysisStep, data map[string]interface{}, callback StreamCallback) error {
	systemPrompt := GetSystemPrompt(step)
	userPrompt := BuildUserPrompt(step, data)

	reqBody := deepSeekRequest{
		Model: d.model,
		Messages: []deepSeekMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Stream:      true,
		Temperature: 0.7,
		MaxTokens:   800,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", d.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.apiKey)

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("DeepSeek API错误 (状态码: %d): %s", resp.StatusCode, string(body))
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("读取流失败: %w", err)
		}

		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		if bytes.HasPrefix(line, []byte("data: ")) {
			line = bytes.TrimPrefix(line, []byte("data: "))
		}

		if bytes.Equal(line, []byte("[DONE]")) {
			break
		}

		var streamResp deepSeekStreamResponse
		if err := json.Unmarshal(line, &streamResp); err != nil {
			continue
		}

		if len(streamResp.Choices) > 0 {
			content := streamResp.Choices[0].Delta.Content
			if content != "" {
				if err := callback(content); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
