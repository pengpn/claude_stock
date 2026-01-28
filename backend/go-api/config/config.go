package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	PythonServiceURL  string
	ClaudeAPIKey      string
}

var AppConfig *Config

func Load() {
	// 加载.env文件
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("未找到.env文件，使用环境变量")
	}

	AppConfig = &Config{
		Port:              getEnv("GO_API_PORT", "8080"),
		PythonServiceURL:  getEnv("PYTHON_SERVICE_URL", "http://localhost:5000"),
		ClaudeAPIKey:      getEnv("CLAUDE_API_KEY", ""),
	}

	if AppConfig.ClaudeAPIKey == "" {
		log.Fatal("CLAUDE_API_KEY未配置")
	}

	log.Printf("配置加载完成 - Port: %s, Python: %s", AppConfig.Port, AppConfig.PythonServiceURL)
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
