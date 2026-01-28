package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"stock-analysis-api/backend/go-api/config"
	"stock-analysis-api/backend/go-api/internal/model"
	"time"
)

type PythonClient struct {
	baseURL string
	client  *http.Client
}

func NewPythonClient() *PythonClient {
	return &PythonClient{
		baseURL: config.AppConfig.PythonServiceURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Analyze 调用Python分析服务
func (pc *PythonClient) Analyze(code string) (*model.PythonAnalysisResponse, error) {
	url := pc.baseURL + "/analyze"

	reqBody := map[string]string{"code": code}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	resp, err := pc.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("调用Python服务失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Python服务返回错误: %d - %s", resp.StatusCode, string(body))
	}

	var result model.PythonAnalysisResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &result, nil
}
