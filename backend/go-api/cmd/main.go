package main

import (
	"log"
	"stock-analysis-api/backend/go-api/config"
	"stock-analysis-api/backend/go-api/internal/client"

	"github.com/gin-gonic/gin"
)

func main() {
	config.Load()

	r := gin.Default()

	// 初始化Python客户端
	pythonClient := client.NewPythonClient()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "go-api"})
	})

	// 测试Python连接
	r.GET("/test-python/:code", func(c *gin.Context) {
		code := c.Param("code")
		result, err := pythonClient.Analyze(code)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, result)
	})

	addr := ":" + config.AppConfig.Port
	log.Printf("Go API服务启动在 %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal("启动失败:", err)
	}
}
