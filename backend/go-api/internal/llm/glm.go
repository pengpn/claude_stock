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
	"strings"
)

type GLMClient struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// GLM API 请求/响应结构
type glmMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type glmRequest struct {
	Model       string       `json:"model"`
	Messages    []glmMessage `json:"messages"`
	Temperature float64      `json:"temperature"`
	MaxTokens   int          `json:"max_tokens,omitempty"`
	Stream      bool         `json:"stream"`
}

type glmStreamResponse struct {
	ID      string `json:"id"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

func NewGLMClient() *GLMClient {
	baseURL := config.AppConfig.GLMBaseURL
	if baseURL == "" {
		baseURL = "https://open.bigmodel.cn/api/paas/v4"
	}

	model := config.AppConfig.GLMModel
	if model == "" {
		model = "glm-4-plus" // 默认使用 GLM-4-Plus
	}

	return &GLMClient{
		apiKey:  config.AppConfig.GLMAPIKey,
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{},
	}
}

func (g *GLMClient) StreamAnalyze(ctx context.Context, step AnalysisStep, data map[string]interface{}, callback StreamCallback) error {
	systemPrompt := GetSystemPrompt(step)
	userPrompt := BuildUserPrompt(step, data)

	// 构建请求
	reqBody := glmRequest{
		Model: g.model,
		Messages: []glmMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
		Temperature: 0.7,
		MaxTokens:   800,
		Stream:      true,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("JSON序列化失败: %w", err)
	}

	// 创建HTTP请求
	url := fmt.Sprintf("%s/chat/completions", g.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", g.apiKey))

	// 发送请求
	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GLM API错误 (状态码 %d): %s", resp.StatusCode, string(bodyBytes))
	}

	// 处理流式响应
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("读取响应失败: %w", err)
		}

		// 跳过空行
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		// SSE格式: data: {...}
		if !bytes.HasPrefix(line, []byte("data: ")) {
			continue
		}

		data := bytes.TrimPrefix(line, []byte("data: "))

		// 检查是否是结束标记
		if string(data) == "[DONE]" {
			break
		}

		// 解析JSON
		var streamResp glmStreamResponse
		if err := json.Unmarshal(data, &streamResp); err != nil {
			// 忽略解析错误，继续处理下一行
			continue
		}

		// 提取内容增量
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

// ValidateConfig 验证GLM配置
func (g *GLMClient) ValidateConfig() error {
	if g.apiKey == "" {
		return fmt.Errorf("GLM API密钥未配置")
	}
	if !strings.HasPrefix(g.apiKey, "sk-") && len(g.apiKey) < 32 {
		return fmt.Errorf("GLM API密钥格式无效")
	}
	return nil
}
