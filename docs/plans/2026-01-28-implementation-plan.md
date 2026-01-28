# 智能股票分析小程序实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 实现一个完整的股票AI分析小程序，包含Python数据分析服务、Go API网关和Vue小程序前端

**Architecture:**
- Python Flask服务提供股票数据分析（akshare获取数据 + 财务指标计算）
- Go Gin服务作为API网关，编排AI分析流程，通过SSE流式返回结果
- uni-app小程序前端实时展示5步AI分析（综合分析、多空辩论、交易员决策、最终决策）

**Tech Stack:**
- Backend: Go 1.20+ (Gin), Python 3.10+ (Flask, akshare, pandas)
- Frontend: Vue 3, uni-app, TypeScript
- AI: Claude API (Anthropic SDK)
- Local Development: localhost无需部署

---

## Phase 1: 项目初始化和基础设施

### Task 1.1: 创建项目目录结构

**Files:**
- Create: `backend/python-analysis/.gitkeep`
- Create: `backend/go-api/.gitkeep`
- Create: `frontend/miniapp/.gitkeep`
- Create: `.gitignore`
- Create: `.env.example`
- Create: `README.md`

**Step 1: 创建目录结构**

```bash
mkdir -p backend/python-analysis/{services,utils}
mkdir -p backend/go-api/{cmd,internal/{handler,service,llm,client,model},config}
mkdir -p frontend/miniapp/{pages/index/components,api,utils,static}
```

**Step 2: 创建 .gitignore**

```
# Python
__pycache__/
*.py[cod]
*$py.class
.venv/
venv/
*.egg-info/
.cache/

# Go
*.exe
*.exe~
*.dll
*.so
*.dylib
vendor/

# Frontend
node_modules/
dist/
unpackage/
*.log

# Environment
.env
.DS_Store

# IDE
.vscode/
.idea/
```

**Step 3: 创建 .env.example**

```bash
# Claude API配置
CLAUDE_API_KEY=sk-ant-your-api-key-here

# Python服务地址
PYTHON_SERVICE_URL=http://localhost:5000

# 服务端口
GO_API_PORT=8080
PYTHON_API_PORT=5000
```

**Step 4: 创建 README.md**

```markdown
# 智能股票分析小程序

基于AI的股票投资分析工具，提供多角色分析和投资建议。

## 技术栈
- Backend: Go (Gin) + Python (Flask)
- Frontend: Vue 3 + uni-app
- AI: Claude API

## 本地开发

### 1. 启动Python分析服务
```bash
cd backend/python-analysis
pip install -r requirements.txt
python app.py
```

### 2. 启动Go API服务
```bash
cd backend/go-api
cp ../../.env.example .env  # 配置CLAUDE_API_KEY
go run cmd/main.go
```

### 3. 启动小程序
```bash
cd frontend/miniapp
npm install
npm run dev:mp-weixin
# 在微信开发者工具中导入 dist/dev/mp-weixin
```

## License
MIT
```

**Step 5: 提交初始化**

```bash
git add .
git commit -m "Init: 项目结构初始化

- 创建backend/frontend目录结构
- 添加.gitignore和环境变量模板
- 添加README文档

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Phase 2: Python分析服务实现

### Task 2.1: Python服务基础设施

**Files:**
- Create: `backend/python-analysis/requirements.txt`
- Create: `backend/python-analysis/config.py`
- Create: `backend/python-analysis/app.py`
- Create: `backend/python-analysis/utils/logger.py`

**Step 1: 创建 requirements.txt**

```txt
flask==3.0.0
flask-cors==4.0.0
akshare==1.18.20
pandas==2.1.4
numpy==1.26.2
python-dotenv==1.0.0
```

**Step 2: 创建 config.py**

```python
import os
from dotenv import load_dotenv

load_dotenv()

class Config:
    PORT = int(os.getenv('PYTHON_API_PORT', 5000))
    DEBUG = os.getenv('DEBUG', 'False').lower() == 'true'

config = Config()
```

**Step 3: 创建 utils/logger.py**

```python
import logging
import sys

def setup_logger(name: str) -> logging.Logger:
    logger = logging.getLogger(name)
    logger.setLevel(logging.INFO)

    handler = logging.StreamHandler(sys.stdout)
    handler.setLevel(logging.INFO)

    formatter = logging.Formatter(
        '%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )
    handler.setFormatter(formatter)

    logger.addHandler(handler)
    return logger

logger = setup_logger('stock-analysis')
```

**Step 4: 创建基础 app.py**

```python
from flask import Flask, jsonify
from flask_cors import CORS
from config import config
from utils.logger import logger

app = Flask(__name__)
CORS(app)

@app.route('/health', methods=['GET'])
def health():
    return jsonify({"status": "ok", "service": "python-analysis"})

if __name__ == '__main__':
    logger.info(f"Starting Python Analysis Service on port {config.PORT}")
    app.run(host='0.0.0.0', port=config.PORT, debug=config.DEBUG)
```

**Step 5: 测试基础服务**

```bash
cd backend/python-analysis
pip install -r requirements.txt
python app.py
```

期望输出：
```
Starting Python Analysis Service on port 5000
 * Running on http://0.0.0.0:5000
```

测试health接口：
```bash
curl http://localhost:5000/health
```

期望返回：
```json
{"status":"ok","service":"python-analysis"}
```

**Step 6: 提交**

```bash
git add backend/python-analysis/
git commit -m "Add: Python分析服务基础框架

- Flask应用初始化
- 配置管理和日志工具
- Health check接口

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

### Task 2.2: 股票数据获取服务

**Files:**
- Create: `backend/python-analysis/services/data_fetcher.py`

**Step 1: 创建 data_fetcher.py**

```python
import akshare as ak
import pandas as pd
from typing import Optional, Dict, Any
from utils.logger import logger

class StockDataFetcher:
    """股票数据获取服务"""

    def __init__(self):
        self.logger = logger

    def safe_float(self, value) -> Optional[float]:
        """安全转换为浮点数"""
        if value is None or value == '' or value == '--':
            return None
        try:
            if pd.isna(value):
                return None
            if isinstance(value, str):
                value = value.replace('%', '').replace(',', '').replace('亿', '')
            return float(value)
        except (ValueError, TypeError):
            return None

    def get_basic_info(self, code: str) -> Dict[str, Any]:
        """获取股票基本信息"""
        try:
            df = ak.stock_individual_info_em(symbol=code)
            info = {}
            for _, row in df.iterrows():
                info[row['item']] = row['value']

            return {
                "code": code,
                "name": info.get("股票简称", ""),
                "industry": info.get("行业", ""),
                "market_cap": self.safe_float(info.get("总市值")),
                "pe_ttm": self.safe_float(info.get("市盈率(动态)")),
                "pb": self.safe_float(info.get("市净率")),
                "listing_date": info.get("上市时间", "")
            }
        except Exception as e:
            self.logger.error(f"获取基本信息失败: {e}")
            return {"code": code, "error": str(e)}

    def get_latest_price(self, code: str) -> Dict[str, Any]:
        """获取最新价格数据"""
        try:
            df = ak.stock_zh_a_hist(
                symbol=code,
                period="daily",
                adjust="qfq"
            )
            if df is None or df.empty:
                return {"error": "无价格数据"}

            latest = df.iloc[-1]
            return {
                "latest_price": self.safe_float(latest['收盘']),
                "price_change_pct": self.safe_float(latest['涨跌幅']),
                "date": str(latest['日期'])
            }
        except Exception as e:
            self.logger.error(f"获取价格数据失败: {e}")
            return {"error": str(e)}

    def get_financial_indicators(self, code: str, limit: int = 4) -> list:
        """获取财务指标"""
        try:
            df = ak.stock_financial_analysis_indicator(symbol=code)
            if df is None or df.empty:
                return []
            return df.head(limit).to_dict(orient='records')
        except Exception as e:
            self.logger.error(f"获取财务指标失败: {e}")
            return []

    def fetch_all(self, code: str) -> Dict[str, Any]:
        """获取所有数据"""
        self.logger.info(f"开始获取股票数据: {code}")

        result = {
            "code": code,
            "basic_info": self.get_basic_info(code),
            "price": self.get_latest_price(code),
            "financial_indicators": self.get_financial_indicators(code)
        }

        self.logger.info(f"数据获取完成: {code}")
        return result
```

**Step 2: 添加测试脚本**

在 `backend/python-analysis/` 创建 `test_fetch.py`:

```python
from services.data_fetcher import StockDataFetcher
import json

if __name__ == '__main__':
    fetcher = StockDataFetcher()

    # 测试铜陵有色
    result = fetcher.fetch_all("000630")
    print(json.dumps(result, ensure_ascii=False, indent=2))
```

**Step 3: 测试数据获取**

```bash
cd backend/python-analysis
python test_fetch.py
```

期望输出包含基本信息、价格、财务指标

**Step 4: 提交**

```bash
git add backend/python-analysis/services/
git commit -m "Add: 股票数据获取服务

- 实现StockDataFetcher类
- 支持获取基本信息、价格、财务指标
- 使用akshare数据源

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

### Task 2.3: 财务分析服务

**Files:**
- Create: `backend/python-analysis/services/financial_analyzer.py`

**Step 1: 创建 financial_analyzer.py**

```python
from typing import Dict, Any, List
from utils.logger import logger

class FinancialAnalyzer:
    """财务分析服务"""

    def __init__(self):
        self.logger = logger

    def extract_latest_metrics(self, indicators: list) -> Dict[str, Any]:
        """提取最新财务指标"""
        if not indicators:
            return {}

        latest = indicators[0]

        return {
            "roe": self._safe_get(latest, "净资产收益率", "ROE"),
            "roa": self._safe_get(latest, "总资产净利率", "ROA"),
            "gross_margin": self._safe_get(latest, "销售毛利率", "毛利率"),
            "net_margin": self._safe_get(latest, "销售净利率", "净利率"),
            "debt_ratio": self._safe_get(latest, "资产负债率", "负债率"),
            "current_ratio": self._safe_get(latest, "流动比率"),
            "quick_ratio": self._safe_get(latest, "速动比率"),
        }

    def _safe_get(self, data: dict, *keys) -> float:
        """安全获取字典值"""
        for key in keys:
            if key in data:
                val = data[key]
                if val and val != '--':
                    try:
                        return float(str(val).replace('%', ''))
                    except:
                        pass
        return 0.0

    def calculate_growth(self, indicators: list) -> Dict[str, float]:
        """计算增长率"""
        if len(indicators) < 2:
            return {"revenue_growth": 0.0, "profit_growth": 0.0}

        latest = indicators[0]
        previous = indicators[1]

        revenue_growth = self._calc_growth_rate(
            self._safe_get(latest, "营业总收入"),
            self._safe_get(previous, "营业总收入")
        )

        profit_growth = self._calc_growth_rate(
            self._safe_get(latest, "净利润"),
            self._safe_get(previous, "净利润")
        )

        return {
            "revenue_growth": revenue_growth,
            "profit_growth": profit_growth
        }

    def _calc_growth_rate(self, current: float, previous: float) -> float:
        """计算增长率"""
        if previous == 0 or previous is None:
            return 0.0
        return ((current - previous) / previous) * 100

    def detect_risks(self, metrics: Dict[str, Any], basic_info: Dict[str, Any]) -> List[str]:
        """检测财务风险"""
        risks = []

        # 高负债风险
        debt_ratio = metrics.get("debt_ratio", 0)
        if debt_ratio > 70:
            risks.append("资产负债率过高")
        elif debt_ratio > 60:
            risks.append("资产负债率偏高")

        # 低盈利能力
        roe = metrics.get("roe", 0)
        if roe < 5:
            risks.append("净资产收益率较低")

        # 流动性风险
        current_ratio = metrics.get("current_ratio", 0)
        if current_ratio < 1:
            risks.append("流动比率低于1，短期偿债能力弱")

        # 估值风险
        pe = basic_info.get("pe_ttm", 0)
        if pe and pe > 50:
            risks.append("市盈率较高，估值偏贵")

        return risks if risks else ["未检测到明显风险"]

    def analyze(self, stock_data: Dict[str, Any]) -> Dict[str, Any]:
        """综合分析"""
        self.logger.info(f"开始财务分析: {stock_data.get('code')}")

        indicators = stock_data.get("financial_indicators", [])
        basic_info = stock_data.get("basic_info", {})

        metrics = self.extract_latest_metrics(indicators)
        growth = self.calculate_growth(indicators)
        risks = self.detect_risks(metrics, basic_info)

        return {
            "financial_metrics": {**metrics, **growth},
            "risks": risks
        }
```

**Step 2: 更新 app.py 添加分析接口**

在 `backend/python-analysis/app.py` 中添加：

```python
from flask import Flask, jsonify, request
from flask_cors import CORS
from config import config
from utils.logger import logger
from services.data_fetcher import StockDataFetcher
from services.financial_analyzer import FinancialAnalyzer

app = Flask(__name__)
CORS(app)

# 初始化服务
data_fetcher = StockDataFetcher()
financial_analyzer = FinancialAnalyzer()

@app.route('/health', methods=['GET'])
def health():
    return jsonify({"status": "ok", "service": "python-analysis"})

@app.route('/analyze', methods=['POST'])
def analyze():
    try:
        data = request.get_json()
        code = data.get('code')

        if not code:
            return jsonify({"error": "缺少股票代码"}), 400

        # 获取数据
        stock_data = data_fetcher.fetch_all(code)

        if "error" in stock_data.get("basic_info", {}):
            return jsonify({"error": "股票代码不存在或数据获取失败"}), 404

        # 财务分析
        analysis = financial_analyzer.analyze(stock_data)

        # 合并结果
        result = {
            "code": code,
            "name": stock_data["basic_info"].get("name", ""),
            "basic_info": stock_data["basic_info"],
            "price": stock_data["price"],
            **analysis
        }

        return jsonify(result)

    except Exception as e:
        logger.error(f"分析失败: {e}")
        return jsonify({"error": str(e)}), 500

if __name__ == '__main__':
    logger.info(f"Starting Python Analysis Service on port {config.PORT}")
    app.run(host='0.0.0.0', port=config.PORT, debug=config.DEBUG)
```

**Step 3: 测试分析接口**

启动服务：
```bash
cd backend/python-analysis
python app.py
```

测试接口：
```bash
curl -X POST http://localhost:5000/analyze \
  -H "Content-Type: application/json" \
  -d '{"code":"000630"}'
```

期望返回包含基本信息、财务指标、风险分析的JSON

**Step 4: 提交**

```bash
git add backend/python-analysis/
git commit -m "Add: 财务分析服务

- 实现FinancialAnalyzer类
- 提取财务指标、计算增长率
- 财务风险检测
- 添加/analyze接口

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Phase 3: Go API服务实现

### Task 3.1: Go项目初始化

**Files:**
- Create: `backend/go-api/go.mod`
- Create: `backend/go-api/cmd/main.go`
- Create: `backend/go-api/config/config.go`

**Step 1: 初始化Go模块**

```bash
cd backend/go-api
go mod init stock-analysis-api
```

**Step 2: 安装依赖**

```bash
go get github.com/gin-gonic/gin@v1.9.1
go get github.com/joho/godotenv@v1.5.1
go get github.com/anthropics/anthropic-sdk-go@latest
```

**Step 3: 创建 config/config.go**

```go
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
```

**Step 4: 创建 cmd/main.go**

```go
package main

import (
	"log"
	"stock-analysis-api/config"

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
```

**Step 5: 测试Go服务**

```bash
cd backend/go-api
go run cmd/main.go
```

期望输出：
```
配置加载完成 - Port: 8080, Python: http://localhost:5000
Go API服务启动在 :8080
```

测试：
```bash
curl http://localhost:8080/health
```

**Step 6: 提交**

```bash
git add backend/go-api/
git commit -m "Add: Go API服务框架

- 初始化Go模块和依赖
- 配置管理
- 基础Gin服务和健康检查

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

### Task 3.2: Python客户端

**Files:**
- Create: `backend/go-api/internal/model/types.go`
- Create: `backend/go-api/internal/client/python_client.go`

**Step 1: 创建 internal/model/types.go**

```go
package model

// StockAnalyzeRequest 股票分析请求
type StockAnalyzeRequest struct {
	Code string `json:"code" binding:"required"`
	Name string `json:"name"`
}

// BasicInfo 基本信息
type BasicInfo struct {
	Code       string  `json:"code"`
	Name       string  `json:"name"`
	Industry   string  `json:"industry"`
	MarketCap  float64 `json:"market_cap"`
	PETTM      float64 `json:"pe_ttm"`
	PB         float64 `json:"pb"`
}

// PriceInfo 价格信息
type PriceInfo struct {
	LatestPrice     float64 `json:"latest_price"`
	PriceChangePct  float64 `json:"price_change_pct"`
	Date            string  `json:"date"`
}

// FinancialMetrics 财务指标
type FinancialMetrics struct {
	ROE            float64 `json:"roe"`
	ROA            float64 `json:"roa"`
	GrossMargin    float64 `json:"gross_margin"`
	NetMargin      float64 `json:"net_margin"`
	DebtRatio      float64 `json:"debt_ratio"`
	CurrentRatio   float64 `json:"current_ratio"`
	RevenueGrowth  float64 `json:"revenue_growth"`
	ProfitGrowth   float64 `json:"profit_growth"`
}

// PythonAnalysisResponse Python分析响应
type PythonAnalysisResponse struct {
	Code              string           `json:"code"`
	Name              string           `json:"name"`
	BasicInfo         BasicInfo        `json:"basic_info"`
	Price             PriceInfo        `json:"price"`
	FinancialMetrics  FinancialMetrics `json:"financial_metrics"`
	Risks             []string         `json:"risks"`
}
```

**Step 2: 创建 internal/client/python_client.go**

```go
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"stock-analysis-api/config"
	"stock-analysis-api/internal/model"
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
```

**Step 3: 添加测试接口到 cmd/main.go**

```go
package main

import (
	"log"
	"stock-analysis-api/config"
	"stock-analysis-api/internal/client"

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
```

**Step 4: 测试Python客户端**

确保Python服务在运行，然后：

```bash
cd backend/go-api
go run cmd/main.go
```

测试：
```bash
curl http://localhost:8080/test-python/000630
```

期望返回Python分析的完整JSON数据

**Step 5: 提交**

```bash
git add backend/go-api/
git commit -m "Add: Python服务客户端

- 定义数据模型
- 实现PythonClient调用分析接口
- 添加测试接口验证连通性

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

### Task 3.3: Claude LLM抽象层

**Files:**
- Create: `backend/go-api/internal/llm/client.go`
- Create: `backend/go-api/internal/llm/claude.go`
- Create: `backend/go-api/internal/llm/prompts.go`

**Step 1: 创建 internal/llm/client.go**

```go
package llm

import "context"

// AnalysisStep 分析步骤类型
type AnalysisStep string

const (
	StepComprehensive AnalysisStep = "comprehensive"
	StepDebateBull    AnalysisStep = "debate_bull"
	StepDebateBear    AnalysisStep = "debate_bear"
	StepTrader        AnalysisStep = "trader"
	StepFinal         AnalysisStep = "final"
)

// StreamCallback 流式响应回调
type StreamCallback func(content string) error

// LLMClient LLM客户端接口
type LLMClient interface {
	// StreamAnalyze 流式分析
	StreamAnalyze(ctx context.Context, step AnalysisStep, data map[string]interface{}, callback StreamCallback) error
}
```

**Step 2: 创建 internal/llm/prompts.go**

```go
package llm

import "fmt"

// GetSystemPrompt 获取系统提示词
func GetSystemPrompt(step AnalysisStep) string {
	prompts := map[AnalysisStep]string{
		StepComprehensive: `你是一位资深的A股投资分析师，擅长客观中立地分析上市公司。请基于提供的财务数据进行综合分析，包括：
1. 公司基本情况和行业地位
2. 财务健康度（盈利能力、偿债能力）
3. 估值水平评估
4. 主要风险点

要求：客观中立，基于数据，200-300字，结构清晰。`,

		StepDebateBull: `你是一位乐观的多头投资者，擅长挖掘股票的投资价值和上涨潜力。请从多头角度分析：
1. 最吸引人的3-5个投资亮点
2. 为什么现在是好的买入时机
3. 未来的上涨驱动力

要求：积极正面但基于数据，150-200字，突出投资价值。`,

		StepDebateBear: `你是一位谨慎的空头投资者，擅长识别风险和质疑过度乐观的预期。请从空头角度分析：
1. 最大的3-5个风险点
2. 为什么当前估值可能不便宜
3. 哪些因素可能导致下跌

要求：批判谨慎但基于逻辑，150-200字，突出风险因素。`,

		StepTrader: `你是一位实战经验丰富的A股交易员，擅长将分析转化为具体的交易决策。基于前面的多空分析，给出：
1. 操作方向（买入/持有/卖出）
2. 建议仓位（轻仓5-10%/中仓10-20%/重仓20%+）
3. 参考买入价位区间
4. 止损位设置
5. 预期持有周期

要求：具体可执行，考虑风险收益比，150-200字。`,

		StepFinal: `你是投资决策委员会的风险管理官，负责综合各方意见给出最终决策。请提供：
1. 风险等级评估（高/中/低风险）
2. 综合投资建议（买入/持有/卖出）
3. 信心指数（0-100）
4. 决策理由总结

要求：平衡风险和收益，给出明确结论，200-250字。`,
	}

	return prompts[step]
}

// BuildUserPrompt 构建用户提示词
func BuildUserPrompt(step AnalysisStep, data map[string]interface{}) string {
	name := data["name"].(string)
	code := data["code"].(string)

	switch step {
	case StepComprehensive:
		return fmt.Sprintf(`请分析【%s(%s)】：

【基本信息】
- 行业: %v
- 市值: %.2f亿元
- 最新价: %.2f元
- PE: %.2f, PB: %.2f

【财务指标】
- ROE: %.2f%%
- 资产负债率: %.2f%%
- 营收增长: %.2f%%
- 净利润增长: %.2f%%

【风险信号】
%v

请进行综合分析。`,
			name, code,
			data["industry"],
			data["market_cap"],
			data["latest_price"],
			data["pe_ttm"],
			data["pb"],
			data["roe"],
			data["debt_ratio"],
			data["revenue_growth"],
			data["profit_growth"],
			data["risks"])

	case StepDebateBull, StepDebateBear:
		previous := data["comprehensive_analysis"].(string)
		return fmt.Sprintf(`基于以下综合分析，请给出【%s】的看%s观点：

【综合分析】
%s

【关键数据】
- ROE: %.2f%%
- 资产负债率: %.2f%%
- 营收增长: %.2f%%

请从%s角度分析。`,
			name,
			map[AnalysisStep]string{StepDebateBull: "多", StepDebateBear: "空"}[step],
			previous,
			data["roe"],
			data["debt_ratio"],
			data["revenue_growth"],
			map[AnalysisStep]string{StepDebateBull: "多头", StepDebateBear: "空头"}[step])

	case StepTrader:
		return fmt.Sprintf(`基于以下分析，给出【%s】的交易建议：

【综合分析】
%s

【多头观点】
%s

【空头观点】
%s

【当前价格】%.2f元

请给出具体的交易建议。`,
			name,
			data["comprehensive_analysis"],
			data["bull_case"],
			data["bear_case"],
			data["latest_price"])

	case StepFinal:
		return fmt.Sprintf(`基于完整分析链，给出【%s】的最终投资建议：

【综合分析】
%s

【多头观点】
%s

【空头观点】
%s

【交易员建议】
%s

请给出最终决策（包含：风险等级、投资建议、信心指数、理由）。`,
			name,
			data["comprehensive_analysis"],
			data["bull_case"],
			data["bear_case"],
			data["trader_decision"])

	default:
		return "请进行分析"
	}
}
```

**Step 3: 创建 internal/llm/claude.go**

```go
package llm

import (
	"context"
	"fmt"
	"stock-analysis-api/config"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type ClaudeClient struct {
	client *anthropic.Client
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
		Model:     anthropic.F(anthropic.ModelClaude3_5Sonnet20241022),
		MaxTokens: anthropic.Int(800),
		System: anthropic.F([]anthropic.TextBlockParam{
			anthropic.NewTextBlock(systemPrompt),
		}),
		Messages: anthropic.F([]anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userPrompt)),
		}),
		Temperature: anthropic.Float(0.7),
	})

	// 处理流式响应
	for stream.Next() {
		event := stream.Current()

		// 处理内容增量
		if delta, ok := event.Delta.(anthropic.ContentBlockDeltaEventDelta); ok {
			if textDelta, ok := delta.AsUnion().(anthropic.ContentBlockDeltaEventDeltaTextDelta); ok {
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
```

**Step 4: 提交**

```bash
git add backend/go-api/internal/llm/
git commit -m "Add: Claude LLM抽象层

- 定义LLMClient接口
- 实现ClaudeClient流式调用
- 5个分析步骤的Prompt模板

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

### Task 3.4: 分析服务编排

**Files:**
- Create: `backend/go-api/internal/service/orchestrator.go`
- Create: `backend/go-api/internal/handler/analyze.go`

**Step 1: 创建 internal/service/orchestrator.go**

```go
package service

import (
	"context"
	"fmt"
	"log"
	"stock-analysis-api/internal/client"
	"stock-analysis-api/internal/llm"
	"stock-analysis-api/internal/model"
)

// SSEEvent SSE事件
type SSEEvent struct {
	Event string
	Data  interface{}
}

// AnalysisOrchestrator 分析编排器
type AnalysisOrchestrator struct {
	pythonClient *client.PythonClient
	llmClient    llm.LLMClient
}

func NewAnalysisOrchestrator(pythonClient *client.PythonClient, llmClient llm.LLMClient) *AnalysisOrchestrator {
	return &AnalysisOrchestrator{
		pythonClient: pythonClient,
		llmClient:    llmClient,
	}
}

// Analyze 执行完整分析流程
func (ao *AnalysisOrchestrator) Analyze(ctx context.Context, code string, eventChan chan<- SSEEvent) error {
	defer close(eventChan)

	// 步骤0: 获取Python分析数据
	eventChan <- SSEEvent{
		Event: "progress",
		Data: map[string]interface{}{
			"step":     "fetching_data",
			"message":  "正在获取股票数据...",
			"progress": 10,
		},
	}

	pythonData, err := ao.pythonClient.Analyze(code)
	if err != nil {
		return fmt.Errorf("获取数据失败: %w", err)
	}

	// 准备LLM输入数据
	llmData := ao.prepareLLMData(pythonData)

	// 存储各步骤结果
	results := make(map[string]string)

	// 步骤1: 综合分析
	if err := ao.runStep(ctx, llm.StepComprehensive, "综合分析", llmData, results, eventChan, 20); err != nil {
		return err
	}
	llmData["comprehensive_analysis"] = results[string(llm.StepComprehensive)]

	// 步骤2: 多头观点
	if err := ao.runStep(ctx, llm.StepDebateBull, "多头观点", llmData, results, eventChan, 40); err != nil {
		return err
	}
	llmData["bull_case"] = results[string(llm.StepDebateBull)]

	// 步骤3: 空头观点
	if err := ao.runStep(ctx, llm.StepDebateBear, "空头观点", llmData, results, eventChan, 60); err != nil {
		return err
	}
	llmData["bear_case"] = results[string(llm.StepDebateBear)]

	// 步骤4: 交易员决策
	if err := ao.runStep(ctx, llm.StepTrader, "交易员决策", llmData, results, eventChan, 80); err != nil {
		return err
	}
	llmData["trader_decision"] = results[string(llm.StepTrader)]

	// 步骤5: 最终决策
	if err := ao.runStep(ctx, llm.StepFinal, "最终决策", llmData, results, eventChan, 100); err != nil {
		return err
	}

	// 发送完成事件
	eventChan <- SSEEvent{
		Event: "done",
		Data:  map[string]string{"message": "分析完成"},
	}

	return nil
}

func (ao *AnalysisOrchestrator) runStep(
	ctx context.Context,
	step llm.AnalysisStep,
	stepName string,
	data map[string]interface{},
	results map[string]string,
	eventChan chan<- SSEEvent,
	progress int,
) error {
	log.Printf("开始执行: %s", stepName)

	var content string
	callback := func(delta string) error {
		content += delta
		// 发送流式内容
		eventChan <- SSEEvent{
			Event: "analysis_step",
			Data: map[string]interface{}{
				"step":     string(step),
				"role":     stepName,
				"content":  delta,
				"progress": progress,
			},
		}
		return nil
	}

	if err := ao.llmClient.StreamAnalyze(ctx, step, data, callback); err != nil {
		return fmt.Errorf("%s失败: %w", stepName, err)
	}

	results[string(step)] = content
	log.Printf("完成执行: %s", stepName)
	return nil
}

func (ao *AnalysisOrchestrator) prepareLLMData(pythonData *model.PythonAnalysisResponse) map[string]interface{} {
	return map[string]interface{}{
		"code":            pythonData.Code,
		"name":            pythonData.Name,
		"industry":        pythonData.BasicInfo.Industry,
		"market_cap":      pythonData.BasicInfo.MarketCap,
		"pe_ttm":          pythonData.BasicInfo.PETTM,
		"pb":              pythonData.BasicInfo.PB,
		"latest_price":    pythonData.Price.LatestPrice,
		"roe":             pythonData.FinancialMetrics.ROE,
		"debt_ratio":      pythonData.FinancialMetrics.DebtRatio,
		"revenue_growth":  pythonData.FinancialMetrics.RevenueGrowth,
		"profit_growth":   pythonData.FinancialMetrics.ProfitGrowth,
		"risks":           pythonData.Risks,
	}
}
```

**Step 2: 创建 internal/handler/analyze.go**

```go
package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"stock-analysis-api/internal/model"
	"stock-analysis-api/internal/service"

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
			eventChan <- service.SSEEvent{
				Event: "error",
				Data:  map[string]string{"error": err.Error()},
			}
		}
	}()

	// 流式发送事件
	c.Stream(func(w gin.ResponseWriter) bool {
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
		w.(gin.ResponseWriter).Flush()

		return true
	})
}
```

**Step 3: 更新 cmd/main.go 集成完整流程**

```go
package main

import (
	"log"
	"stock-analysis-api/config"
	"stock-analysis-api/internal/client"
	"stock-analysis-api/internal/handler"
	"stock-analysis-api/internal/llm"
	"stock-analysis-api/internal/service"

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
```

**Step 4: 测试完整流程**

确保Python服务运行，配置好`.env`中的`CLAUDE_API_KEY`：

```bash
cd backend/go-api
go run cmd/main.go
```

测试（使用curl观察SSE流）：
```bash
curl -N -X POST http://localhost:8080/api/v1/analyze \
  -H "Content-Type: application/json" \
  -d '{"code":"000630"}'
```

期望看到流式返回的分析步骤

**Step 5: 提交**

```bash
git add backend/go-api/
git commit -m "Add: 分析流程编排和SSE接口

- 实现AnalysisOrchestrator编排5步分析
- 实现SSE流式响应handler
- 集成完整分析链路

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Phase 4: 前端小程序实现

### Task 4.1: uni-app项目初始化

**Files:**
- Create: `frontend/miniapp/manifest.json`
- Create: `frontend/miniapp/pages.json`
- Create: `frontend/miniapp/main.js`
- Create: `frontend/miniapp/App.vue`
- Create: `frontend/miniapp/package.json`

**Step 1: 创建 package.json**

```json
{
  "name": "stock-analysis-miniapp",
  "version": "1.0.0",
  "scripts": {
    "dev:mp-weixin": "uni -p mp-weixin",
    "build:mp-weixin": "uni build -p mp-weixin"
  },
  "dependencies": {
    "@dcloudio/uni-app": "^3.0.0",
    "@dcloudio/uni-mp-weixin": "^3.0.0",
    "vue": "^3.3.0"
  },
  "devDependencies": {
    "@dcloudio/vite-plugin-uni": "^3.0.0",
    "vite": "^4.0.0"
  }
}
```

**Step 2: 创建 manifest.json**

```json
{
  "name": "智能股票分析",
  "appid": "",
  "description": "AI股票投资分析助手",
  "versionName": "1.0.0",
  "versionCode": "100",
  "transformPx": false,
  "mp-weixin": {
    "appid": "",
    "setting": {
      "urlCheck": false,
      "es6": true,
      "minified": true
    },
    "usingComponents": true
  }
}
```

**Step 3: 创建 pages.json**

```json
{
  "pages": [
    {
      "path": "pages/index/index",
      "style": {
        "navigationBarTitleText": "智能股票分析",
        "navigationBarBackgroundColor": "#1890ff",
        "navigationBarTextStyle": "white"
      }
    }
  ],
  "globalStyle": {
    "navigationBarTextStyle": "white",
    "navigationBarTitleText": "股票分析",
    "navigationBarBackgroundColor": "#1890ff",
    "backgroundColor": "#F8F8F8"
  }
}
```

**Step 4: 创建 main.js**

```javascript
import { createSSRApp } from 'vue'
import App from './App.vue'

export function createApp() {
  const app = createSSRApp(App)
  return {
    app
  }
}
```

**Step 5: 创建 App.vue**

```vue
<script setup>
import { onLaunch } from '@dcloudio/uni-app'

onLaunch(() => {
  console.log('App Launch')
})
</script>

<style>
/* 全局样式 */
page {
  background-color: #f5f5f5;
}
</style>
```

**Step 6: 安装依赖**

```bash
cd frontend/miniapp
npm install
```

**Step 7: 提交**

```bash
git add frontend/miniapp/
git commit -m "Add: uni-app小程序项目初始化

- 配置manifest和pages
- 创建App入口
- 配置开发环境

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

### Task 4.2: API封装

**Files:**
- Create: `frontend/miniapp/api/stock.js`
- Create: `frontend/miniapp/utils/sse.js`

**Step 1: 创建 api/stock.js**

```javascript
const API_BASE = 'http://localhost:8080/api/v1'

export const stockApi = {
  // 分析股票（返回SSE连接的task）
  analyze(code) {
    return {
      url: `${API_BASE}/analyze`,
      data: { code }
    }
  }
}
```

**Step 2: 创建 utils/sse.js (SSE polyfill for 微信小程序)**

```javascript
/**
 * 微信小程序SSE客户端
 * 由于小程序不支持原生EventSource，需要手动实现
 */
export class SSEClient {
  constructor(url, data) {
    this.url = url
    this.data = data
    this.listeners = {}
    this.requestTask = null
  }

  // 监听事件
  addEventListener(event, handler) {
    if (!this.listeners[event]) {
      this.listeners[event] = []
    }
    this.listeners[event].push(handler)
  }

  // 触发事件
  emit(event, data) {
    const handlers = this.listeners[event] || []
    handlers.forEach(handler => handler({ data }))
  }

  // 开始连接
  connect() {
    return new Promise((resolve, reject) => {
      this.requestTask = uni.request({
        url: this.url,
        method: 'POST',
        data: this.data,
        header: {
          'Content-Type': 'application/json'
        },
        enableChunked: true, // 开启分块传输
        success: () => {
          resolve()
        },
        fail: (err) => {
          reject(err)
        }
      })

      // 监听数据接收
      this.requestTask.onChunkReceived((res) => {
        const chunk = this.arrayBufferToString(res.data)
        this.parseSSE(chunk)
      })
    })
  }

  // 解析SSE数据
  parseSSE(text) {
    const lines = text.split('\n')
    let event = 'message'
    let data = ''

    lines.forEach(line => {
      if (line.startsWith('event:')) {
        event = line.substring(6).trim()
      } else if (line.startsWith('data:')) {
        data = line.substring(5).trim()
      } else if (line === '') {
        // 空行表示一条消息结束
        if (data) {
          try {
            const parsedData = JSON.parse(data)
            this.emit(event, parsedData)
          } catch (e) {
            console.error('SSE数据解析失败:', e)
          }
          data = ''
          event = 'message'
        }
      }
    })
  }

  // ArrayBuffer转字符串
  arrayBufferToString(buffer) {
    const uint8Array = new Uint8Array(buffer)
    let str = ''
    for (let i = 0; i < uint8Array.length; i++) {
      str += String.fromCharCode(uint8Array[i])
    }
    return decodeURIComponent(escape(str))
  }

  // 关闭连接
  close() {
    if (this.requestTask) {
      this.requestTask.abort()
    }
  }
}
```

**Step 3: 提交**

```bash
git add frontend/miniapp/api/ frontend/miniapp/utils/
git commit -m "Add: API封装和SSE客户端

- 封装股票分析API
- 实现小程序SSE客户端（分块传输）

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

### Task 4.3: 主页面实现

**Files:**
- Create: `frontend/miniapp/pages/index/index.vue`

**Step 1: 创建 pages/index/index.vue**

```vue
<template>
  <view class="container">
    <!-- 搜索框 -->
    <view class="search-box">
      <input
        v-model="stockCode"
        class="input"
        placeholder="输入股票代码，如: 000630"
        :disabled="analyzing"
      />
      <button
        class="btn-analyze"
        @click="startAnalyze"
        :disabled="analyzing || !stockCode"
      >
        {{ analyzing ? '分析中...' : '开始分析' }}
      </button>
    </view>

    <!-- 进度条 -->
    <view v-if="analyzing" class="progress-bar">
      <view class="progress-inner" :style="{ width: progress + '%' }"></view>
      <text class="progress-text">{{ progress }}%</text>
    </view>

    <!-- 分析结果 -->
    <view v-if="results.length > 0" class="results">
      <!-- 综合分析 -->
      <view
        v-for="(result, index) in results"
        :key="index"
        class="result-card"
        :class="result.step"
      >
        <view class="card-header">
          <text class="card-icon">{{ getIcon(result.step) }}</text>
          <text class="card-title">{{ result.role }}</text>
        </view>
        <view class="card-content">
          <text class="content-text">{{ result.content }}</text>
        </view>
      </view>
    </view>

    <!-- 错误提示 -->
    <view v-if="error" class="error-box">
      <text>{{ error }}</text>
      <button @click="error = ''">关闭</button>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { SSEClient } from '../../utils/sse.js'
import { stockApi } from '../../api/stock.js'

const stockCode = ref('')
const analyzing = ref(false)
const progress = ref(0)
const results = ref([])
const error = ref('')

// 开始分析
const startAnalyze = async () => {
  if (!stockCode.value) {
    uni.showToast({ title: '请输入股票代码', icon: 'none' })
    return
  }

  // 重置状态
  analyzing.value = true
  progress.value = 0
  results.value = []
  error.value = ''

  try {
    const api = stockApi.analyze(stockCode.value)
    const sse = new SSEClient(api.url, api.data)

    // 监听进度事件
    sse.addEventListener('progress', (e) => {
      const data = JSON.parse(e.data)
      progress.value = data.progress
    })

    // 监听分析步骤
    sse.addEventListener('analysis_step', (e) => {
      const data = JSON.parse(e.data)

      // 查找是否已有该步骤
      const existingIndex = results.value.findIndex(r => r.step === data.step)

      if (existingIndex >= 0) {
        // 追加内容（流式）
        results.value[existingIndex].content += data.content
      } else {
        // 新增步骤
        results.value.push({
          step: data.step,
          role: data.role,
          content: data.content
        })
      }

      progress.value = data.progress
    })

    // 监听完成事件
    sse.addEventListener('done', () => {
      analyzing.value = false
      progress.value = 100
      uni.showToast({ title: '分析完成', icon: 'success' })
    })

    // 监听错误事件
    sse.addEventListener('error', (e) => {
      const data = JSON.parse(e.data)
      error.value = data.error
      analyzing.value = false
    })

    // 开始连接
    await sse.connect()

  } catch (err) {
    error.value = '连接失败: ' + err.message
    analyzing.value = false
  }
}

// 获取图标
const getIcon = (step) => {
  const icons = {
    'comprehensive': '📊',
    'debate_bull': '🐂',
    'debate_bear': '🐻',
    'trader': '💼',
    'final': '✅'
  }
  return icons[step] || '📝'
}
</script>

<style scoped>
.container {
  padding: 30rpx;
  min-height: 100vh;
}

.search-box {
  display: flex;
  gap: 20rpx;
  margin-bottom: 30rpx;
}

.input {
  flex: 1;
  height: 80rpx;
  padding: 0 30rpx;
  background: white;
  border-radius: 40rpx;
  border: 2rpx solid #e0e0e0;
}

.btn-analyze {
  width: 200rpx;
  height: 80rpx;
  line-height: 80rpx;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border-radius: 40rpx;
  border: none;
  font-size: 28rpx;
}

.btn-analyze:disabled {
  opacity: 0.6;
}

.progress-bar {
  position: relative;
  height: 40rpx;
  background: #f0f0f0;
  border-radius: 20rpx;
  overflow: hidden;
  margin-bottom: 30rpx;
}

.progress-inner {
  height: 100%;
  background: linear-gradient(90deg, #667eea 0%, #764ba2 100%);
  transition: width 0.3s;
}

.progress-text {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 24rpx;
  color: #333;
  font-weight: bold;
}

.results {
  display: flex;
  flex-direction: column;
  gap: 30rpx;
}

.result-card {
  background: white;
  border-radius: 20rpx;
  padding: 30rpx;
  box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.05);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 15rpx;
  margin-bottom: 20rpx;
  padding-bottom: 20rpx;
  border-bottom: 2rpx solid #f0f0f0;
}

.card-icon {
  font-size: 40rpx;
}

.card-title {
  font-size: 32rpx;
  font-weight: bold;
  color: #333;
}

.card-content {
  line-height: 1.8;
}

.content-text {
  font-size: 28rpx;
  color: #666;
  white-space: pre-wrap;
}

.error-box {
  position: fixed;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 600rpx;
  padding: 40rpx;
  background: white;
  border-radius: 20rpx;
  box-shadow: 0 8rpx 24rpx rgba(0, 0, 0, 0.15);
  text-align: center;
}

.error-box text {
  display: block;
  margin-bottom: 30rpx;
  color: #f5222d;
}

.error-box button {
  width: 200rpx;
  height: 70rpx;
  line-height: 70rpx;
  background: #1890ff;
  color: white;
  border-radius: 35rpx;
  border: none;
}
</style>
```

**Step 2: 编译测试**

```bash
cd frontend/miniapp
npm run dev:mp-weixin
```

编译成功后，在微信开发者工具中：
1. 导入项目：选择 `frontend/miniapp/dist/dev/mp-weixin`
2. 在项目设置中勾选"不校验合法域名"
3. 测试输入股票代码并开始分析

**Step 3: 提交**

```bash
git add frontend/miniapp/pages/
git commit -m "Add: 小程序主页面实现

- 股票代码输入和分析按钮
- SSE流式展示分析结果
- 进度条和错误处理
- 美化UI样式

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Phase 5: 集成测试和文档

### Task 5.1: 创建启动脚本

**Files:**
- Create: `scripts/start-dev.sh`
- Create: `scripts/stop-dev.sh`

**Step 1: 创建 scripts/start-dev.sh**

```bash
#!/bin/bash

echo "=== 启动开发环境 ==="

# 检查.env文件
if [ ! -f .env ]; then
  echo "错误: 未找到.env文件，请从.env.example复制并配置"
  exit 1
fi

# 启动Python服务
echo "1. 启动Python分析服务..."
cd backend/python-analysis
python app.py &
PYTHON_PID=$!
echo "   Python服务 PID: $PYTHON_PID"
cd ../..

sleep 2

# 启动Go服务
echo "2. 启动Go API服务..."
cd backend/go-api
go run cmd/main.go &
GO_PID=$!
echo "   Go服务 PID: $GO_PID"
cd ../..

sleep 2

# 启动前端
echo "3. 启动小程序编译..."
cd frontend/miniapp
npm run dev:mp-weixin &
FRONTEND_PID=$!
echo "   前端编译 PID: $FRONTEND_PID"
cd ../..

# 保存PID
echo "$PYTHON_PID" > .dev.pid
echo "$GO_PID" >> .dev.pid
echo "$FRONTEND_PID" >> .dev.pid

echo ""
echo "=== 开发环境启动成功 ==="
echo "Python服务: http://localhost:5000"
echo "Go API: http://localhost:8080"
echo "请在微信开发者工具中打开: frontend/miniapp/dist/dev/mp-weixin"
echo ""
echo "停止服务: ./scripts/stop-dev.sh"
```

**Step 2: 创建 scripts/stop-dev.sh**

```bash
#!/bin/bash

echo "=== 停止开发环境 ==="

if [ -f .dev.pid ]; then
  while read pid; do
    if kill -0 $pid 2>/dev/null; then
      echo "停止进程 $pid"
      kill $pid
    fi
  done < .dev.pid
  rm .dev.pid
  echo "所有服务已停止"
else
  echo "未找到运行中的服务"
fi
```

**Step 3: 添加执行权限**

```bash
chmod +x scripts/start-dev.sh
chmod +x scripts/stop-dev.sh
```

**Step 4: 提交**

```bash
git add scripts/
git commit -m "Add: 开发环境启动脚本

- 一键启动Python+Go+前端服务
- 一键停止所有服务
- PID管理

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

### Task 5.2: 更新README文档

**Files:**
- Modify: `README.md`

**Step 1: 更新README内容**

```markdown
# 智能股票分析小程序

基于AI的股票投资分析工具，提供5步多角色分析和投资建议。

## 功能特性

- 📊 **综合分析**: 资深分析师视角的全面评估
- 🐂 **多头观点**: 挖掘投资亮点和上涨潜力
- 🐻 **空头观点**: 识别风险和下跌因素
- 💼 **交易员决策**: 具体操作建议和仓位管理
- ✅ **最终决策**: 风险评估和投资建议

## 技术栈

**后端**:
- Go 1.20+ (Gin框架) - API网关和流式响应
- Python 3.10+ (Flask) - 股票数据分析
- akshare - A股数据获取
- Claude API - AI分析引擎

**前端**:
- Vue 3 + uni-app - 微信小程序
- SSE - 流式数据展示

## 快速开始

### 前置要求

- Go >= 1.20
- Python >= 3.10
- Node.js >= 16
- 微信开发者工具
- Claude API Key

### 1. 克隆项目

```bash
git clone <your-repo>
cd stock-analysis-miniapp
```

### 2. 配置环境变量

```bash
cp .env.example .env
# 编辑.env，填入你的CLAUDE_API_KEY
```

### 3. 安装依赖

**Python服务**:
```bash
cd backend/python-analysis
pip install -r requirements.txt
```

**Go服务**:
```bash
cd backend/go-api
go mod download
```

**前端**:
```bash
cd frontend/miniapp
npm install
```

### 4. 启动服务

**方式1: 使用脚本（推荐）**
```bash
./scripts/start-dev.sh
```

**方式2: 手动启动**

终端1 - Python服务:
```bash
cd backend/python-analysis
python app.py
```

终端2 - Go服务:
```bash
cd backend/go-api
go run cmd/main.go
```

终端3 - 前端:
```bash
cd frontend/miniapp
npm run dev:mp-weixin
```

### 5. 打开微信开发者工具

1. 导入项目：`frontend/miniapp/dist/dev/mp-weixin`
2. 设置 → 项目设置 → 勾选"不校验合法域名"
3. 开始调试

### 6. 测试分析

输入股票代码测试：
- 000630 (铜陵有色)
- 600519 (贵州茅台)
- 000858 (五粮液)

## 项目结构

```
stock-analysis-miniapp/
├── backend/
│   ├── go-api/              # Go API服务
│   │   ├── cmd/             # 入口
│   │   ├── internal/        # 业务逻辑
│   │   └── config/          # 配置
│   └── python-analysis/     # Python分析服务
│       ├── services/        # 数据获取和分析
│       └── utils/           # 工具类
├── frontend/
│   └── miniapp/             # uni-app小程序
│       ├── pages/           # 页面
│       ├── api/             # API封装
│       └── utils/           # 工具类
├── docs/
│   └── plans/               # 设计和实施文档
└── scripts/                 # 开发脚本
```

## API文档

### Python分析服务 (Port 5000)

**POST /analyze**
```json
请求: {"code": "000630"}
响应: {
  "code": "000630",
  "name": "铜陵有色",
  "basic_info": {...},
  "financial_metrics": {...},
  "risks": [...]
}
```

### Go API服务 (Port 8080)

**POST /api/v1/analyze**
```json
请求: {"code": "000630"}
响应: SSE流式事件
  - event: progress (进度更新)
  - event: analysis_step (分析步骤)
  - event: done (完成)
```

## 开发指南

### 添加新的分析步骤

1. 在 `backend/go-api/internal/llm/prompts.go` 添加新的prompt
2. 在 `backend/go-api/internal/service/orchestrator.go` 添加步骤编排
3. 前端自动展示新步骤

### 切换LLM模型

实现 `LLMClient` 接口即可：
```go
// internal/llm/client.go
type LLMClient interface {
    StreamAnalyze(ctx context.Context, step AnalysisStep, data map[string]interface{}, callback StreamCallback) error
}
```

## 常见问题

**Q: akshare数据获取失败？**
A: 可能是网络问题或接口限流，稍后重试

**Q: Claude API调用失败？**
A: 检查API Key是否正确，账户余额是否充足

**Q: 小程序SSE无响应？**
A: 确保勾选了"不校验合法域名"，检查后端服务是否正常

## 成本估算

- 单次分析约消耗10K tokens
- Claude API成本约 $0.03-0.05/次
- 月度1000次分析约 $30-50

## 免责声明

本系统提供的所有分析和建议仅供参考，不构成任何投资建议。股市有风险，投资需谨慎。

## License

MIT
```

**Step 2: 提交**

```bash
git add README.md
git commit -m "Update: 完善README文档

- 添加详细的快速开始指南
- 说明项目结构和API
- 常见问题解答

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## 总结

**实施计划已完成，包含：**

✅ Phase 1: 项目初始化和目录结构
✅ Phase 2: Python分析服务（数据获取 + 财务分析）
✅ Phase 3: Go API服务（LLM调用 + SSE流式响应）
✅ Phase 4: 前端小程序（Vue3 + SSE客户端）
✅ Phase 5: 集成测试和文档

**关键里程碑：**
- Task 2.3: Python分析服务完成
- Task 3.4: Go API SSE流式接口完成
- Task 4.3: 小程序主页面完成
- Task 5.2: 完整文档和启动脚本

**下一步建议：**
1. 按Task顺序逐步实施
2. 每个Task完成后运行测试
3. 及时提交代码
4. 最后进行端到端集成测试
