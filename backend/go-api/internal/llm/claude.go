package llm

import (
	"context"
	"fmt"
	"stock-analysis-api/backend/go-api/config"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type ClaudeClient struct {
	client anthropic.Client
}

func NewClaudeClient() *ClaudeClient {
	client := anthropic.NewClient(
		option.WithAPIKey(config.AppConfig.ClaudeAPIKey),
	)
	return &ClaudeClient{client: client}
}

func (c *ClaudeClient) StreamAnalyze(ctx context.Context, step AnalysisStep, data map[string]interface{}, callback StreamCallback) error {
	systemPrompt := GetSystemPrompt(step)
	userPrompt := BuildUserPrompt(step, data)

	stream := c.client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5,
		MaxTokens: 800,
		System: []anthropic.TextBlockParam{
			{
				Text: systemPrompt,
			},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(
				anthropic.NewTextBlock(userPrompt),
			),
		},
		Temperature: anthropic.Float(0.7),
	})

	// 处理流式响应
	for stream.Next() {
		event := stream.Current()

		// 处理内容增量 - 检查事件类型
		if event.Type == "content_block_delta" {
			deltaEvent := event.AsContentBlockDelta()
			textDelta := deltaEvent.Delta.AsTextDelta()
			if textDelta.Text != "" {
				if err := callback(textDelta.Text); err != nil {
					return err
				}
			}
		}
	}

	if err := stream.Err(); err != nil {
		return fmt.Errorf("流式处理失败: %w", err)
	}

	return nil
}
