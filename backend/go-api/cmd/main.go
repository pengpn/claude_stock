package main

import (
	"log"
	"stock-analysis-api/backend/go-api/config"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	config.Load()

	// 创建Gin引擎
	r := gin.Default()

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "go-api",
		})
	})

	// 启动服务
	addr := ":" + config.AppConfig.Port
	log.Printf("Go API服务启动在 %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal("启动失败:", err)
	}
}
