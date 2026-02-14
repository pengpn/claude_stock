package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"stock-analysis-api/backend/go-api/internal/model"
	"stock-analysis-api/backend/go-api/internal/service"

	"github.com/gin-gonic/gin"
)

type AnalyzeHandler struct {
	orchestrator *service.AnalysisOrchestrator
}

func NewAnalyzeHandler(orchestrator *service.AnalysisOrchestrator) *AnalyzeHandler {
	return &AnalyzeHandler{orchestrator: orchestrator}
}

// StreamAnalyze SSE流式分析接口
func (h *AnalyzeHandler) StreamAnalyze(c *gin.Context) {
	var req model.StockAnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 设置SSE响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// 创建事件通道
	eventChan := make(chan service.SSEEvent, 10)

	// 启动分析
	ctx := context.Background()
	go func() {
		if err := h.orchestrator.Analyze(ctx, req.Code, eventChan); err != nil {
			log.Printf("分析失败: %v", err)
			// 错误已经在orchestrator中发送到eventChan，这里只记录日志
		}
	}()

	// 流式发送事件
	c.Stream(func(w io.Writer) bool {
		event, ok := <-eventChan
		if !ok {
			return false
		}

		// 序列化数据
		dataJSON, err := json.Marshal(event.Data)
		if err != nil {
			log.Printf("序列化失败: %v", err)
			return false
		}

		// 发送SSE格式
		fmt.Fprintf(w, "event: %s\n", event.Event)
		fmt.Fprintf(w, "data: %s\n\n", dataJSON)
		c.Writer.Flush()

		return true
	})
}
