package main

import (
	"log"
	"stock-analysis-api/backend/go-api/config"
	"stock-analysis-api/backend/go-api/internal/client"
	"stock-analysis-api/backend/go-api/internal/handler"
	"stock-analysis-api/backend/go-api/internal/llm"
	"stock-analysis-api/backend/go-api/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	config.Load()

	r := gin.Default()

	// 初始化客户端
	pythonClient := client.NewPythonClient()
	llmClient := llm.NewClaudeClient()

	// 初始化服务
	orchestrator := service.NewAnalysisOrchestrator(pythonClient, llmClient)

	// 初始化Handler
	analyzeHandler := handler.NewAnalyzeHandler(orchestrator)

	// 路由
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "go-api"})
	})

	api := r.Group("/api/v1")
	{
		api.POST("/analyze", analyzeHandler.StreamAnalyze)
	}

	addr := ":" + config.AppConfig.Port
	log.Printf("Go API服务启动在 %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal("启动失败:", err)
	}
}
