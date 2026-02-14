package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	PythonServiceURL  string
	LLMProvider       string // "claude", "glm", or "deepseek"
	ClaudeAPIKey      string
	GLMAPIKey         string
	GLMBaseURL        string
	GLMModel          string
	DeepSeekAPIKey    string
	DeepSeekBaseURL   string
	DeepSeekModel     string
}

var AppConfig *Config

func Load() {
	// 加载.env文件
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("未找到.env文件，使用环境变量")
	}

	llmProvider := getEnv("LLM_PROVIDER", "glm")

	AppConfig = &Config{
		Port:              getEnv("GO_API_PORT", "8080"),
		PythonServiceURL:  getEnv("PYTHON_SERVICE_URL", "http://localhost:5000"),
		LLMProvider:       llmProvider,
		ClaudeAPIKey:      getEnv("CLAUDE_API_KEY", ""),
		GLMAPIKey:         getEnv("GLM_API_KEY", ""),
		GLMBaseURL:        getEnv("GLM_BASE_URL", "https://open.bigmodel.cn/api/paas/v4"),
		GLMModel:          getEnv("GLM_MODEL", "glm-4-plus"),
		DeepSeekAPIKey:    getEnv("DEEPSEEK_API_KEY", ""),
		DeepSeekBaseURL:   getEnv("DEEPSEEK_BASE_URL", "https://api.deepseek.com/v1"),
		DeepSeekModel:     getEnv("DEEPSEEK_MODEL", "deepseek-chat"),
	}

	// 验证LLM配置
	switch llmProvider {
	case "claude":
		if AppConfig.ClaudeAPIKey == "" {
			log.Fatal("CLAUDE_API_KEY未配置")
		}
	case "glm":
		if AppConfig.GLMAPIKey == "" {
			log.Fatal("GLM_API_KEY未配置")
		}
	case "deepseek":
		if AppConfig.DeepSeekAPIKey == "" {
			log.Fatal("DEEPSEEK_API_KEY未配置")
		}
	default:
		log.Fatalf("不支持的LLM提供商: %s (支持: claude, glm, deepseek)", llmProvider)
	}

	log.Printf("配置加载完成 - Port: %s, Python: %s, LLM: %s", AppConfig.Port, AppConfig.PythonServiceURL, llmProvider)
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
