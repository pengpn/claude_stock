# 智能股票分析系统

基于AI的股票投资分析工具，提供5步多角色分析和投资建议。支持H5网页端访问。

## 功能特性

- 📊 **综合分析**: 资深分析师视角的全面评估
- 🐂 **多头观点**: 挖掘投资亮点和上涨潜力
- 🐻 **空头观点**: 识别风险和下跌因素
- 💼 **交易员决策**: 具体操作建议和仓位管理
- ✅ **最终决策**: 风险评估和投资建议
- 🔍 **智能搜索**: 支持股票代码或名称输入（如"600519"或"贵州茅台"）
- 🎨 **终端风格**: 独特的命令行界面设计，支持深色/浅色主题切换

## 技术栈

**后端**:
- Go 1.20+ (Gin框架) - API网关和流式响应
- Python 3.10+ (Flask) - 股票数据分析
- akshare - A股数据获取
- DeepSeek/GLM API - AI分析引擎

**前端**:
- Vue 3 + Vite - H5网页应用
- SSE - 实时流式数据展示
- 响应式设计 - 支持桌面和移动端

## 快速开始

### 前置要求

- Go >= 1.20
- Python >= 3.10
- Node.js >= 16
- DeepSeek API Key 或 GLM API Key

### 1. 克隆项目

```bash
git clone https://github.com/pengpn/claude_stock.git
cd claude_stock
```

### 2. 配置环境变量

```bash
cp .env.example .env
# 编辑.env，配置以下参数：
# - LLM_PROVIDER: deepseek 或 glm
# - DEEPSEEK_API_KEY 或 GLM_API_KEY
# - PYTHON_SERVICE_URL: http://localhost:8001
# - GO_API_PORT: 8000
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

**H5前端**:
```bash
cd frontend/h5
npm install
```

### 4. 启动服务

**方式1: 使用脚本（推荐）**
```bash
chmod +x ./scripts/start-dev.sh
./scripts/start-dev.sh
```

启动后访问：
- H5前端: http://localhost:3000
- Go API: http://localhost:8000
- Python服务: http://localhost:8001

**方式2: 手动启动**

终端1 - Python服务:
```bash
cd backend/python-analysis
python app.py
# 服务运行在 http://localhost:8001
```

终端2 - Go服务:
```bash
cd backend/go-api
go run cmd/main.go
# 服务运行在 http://localhost:8000
```

终端3 - H5前端:
```bash
cd frontend/h5
npm run dev
# 服务运行在 http://localhost:3000
```

### 5. 停止服务

```bash
./scripts/stop-dev.sh
```

### 6. 测试分析

在浏览器打开 http://localhost:3000，输入股票代码或名称：
- 600519 或 贵州茅台
- 000001 或 平安银行
- 600036 或 招商银行

## 项目结构

```
claude_stock/
├── backend/
│   ├── go-api/              # Go API服务 (端口8000)
│   │   ├── cmd/             # 入口
│   │   ├── internal/        # 业务逻辑
│   │   │   ├── handler/     # HTTP处理器
│   │   │   ├── service/     # 业务服务
│   │   │   ├── client/      # 外部客户端
│   │   │   └── llm/         # LLM集成 (DeepSeek/GLM)
│   │   └── config/          # 配置
│   └── python-analysis/     # Python分析服务 (端口8001)
│       ├── services/        # 数据获取和分析
│       └── utils/           # 工具类
├── frontend/
│   └── h5/                  # Vue3 H5应用 (端口3000)
│       ├── src/             # 源代码
│       │   ├── App.vue      # 主应用组件
│       │   └── style.css    # 全局样式
│       └── index.html       # 入口HTML
├── docs/                    # 技术文档
└── scripts/                 # 开发脚本
    ├── start-dev.sh         # 启动开发环境
    └── stop-dev.sh          # 停止开发环境
```

## API文档

### Python分析服务 (Port 8001)

**POST /analyze**
```json
请求: {"code": "600519"} 或 {"code": "贵州茅台"}
响应: {
  "code": "600519",
  "name": "贵州茅台",
  "basic_info": {
    "industry": "白酒",
    "market_cap": 2500000000000,
    "pe_ttm": 35.5,
    "pb": 12.8
  },
  "price": {
    "latest_price": 1680.50
  },
  "financial_metrics": {
    "roe": 0.32,
    "debt_ratio": 0.15,
    "revenue_growth": 0.18,
    "profit_growth": 0.20
  },
  "risks": ["估值偏高", "行业竞争加剧"]
}
```

### Go API服务 (Port 8000)

**POST /api/v1/analyze**
```json
请求: {"code": "600519"} 或 {"code": "贵州茅台"}
响应: SSE流式事件
  - event: progress (进度更新)
    data: {"step": "fetching_data", "message": "正在获取股票数据...", "progress": 10}

  - event: analysis_step (分析步骤流式输出)
    data: {"step": "comprehensive", "role": "综合分析", "content": "...", "progress": 20}

  - event: step_completed (步骤完成)
    data: {"step": "comprehensive", "completed": true}

  - event: done (全部完成)
    data: {"message": "分析完成"}

  - event: error (错误)
    data: {"error": "错误信息"}
```

## 开发指南

### 查看日志

```bash
# Python服务日志
tail -f /tmp/python-service.log

# Go API日志
tail -f /tmp/go-api.log

# H5前端日志
tail -f /tmp/frontend-h5.log
```

### 添加新的分析步骤

1. 在 `backend/go-api/internal/llm/prompts.go` 添加新的prompt
2. 在 `backend/go-api/internal/service/orchestrator.go` 添加步骤编排
3. 前端自动展示新步骤

### 切换LLM提供商

在 `.env` 文件中配置：
```bash
# 使用DeepSeek
LLM_PROVIDER=deepseek
DEEPSEEK_API_KEY=your_key_here

# 或使用GLM
LLM_PROVIDER=glm
GLM_API_KEY=your_key_here
```

支持的LLM接口实现：
```go
// internal/llm/client.go
type LLMClient interface {
    StreamAnalyze(ctx context.Context, step AnalysisStep, data map[string]interface{}, callback StreamCallback) error
}
```

## 常见问题

**Q: 启动脚本权限不足？**
```bash
chmod +x ./scripts/start-dev.sh
chmod +x ./scripts/stop-dev.sh
```

**Q: akshare数据获取失败？**
A: 可能是网络问题或接口限流，稍后重试

**Q: LLM API调用失败？**
A: 检查 `.env` 中的API Key是否正确，账户余额是否充足

**Q: H5页面无法访问？**
A: 确保3000端口未被占用，检查后端服务是否正常启动

**Q: SSE流式输出中断？**
A: 检查Go API日志，确认LLM服务连接正常

**Q: 输入股票名称无法识别？**
A: 确保Python服务正常运行，akshare数据源可访问

## 成本估算

- 单次分析约消耗10-15K tokens
- DeepSeek API成本约 ¥0.01-0.02/次
- GLM API成本约 ¥0.02-0.03/次
- 月度1000次分析约 ¥10-30

## 免责声明

本系统提供的所有分析和建议仅供参考，不构成任何投资建议。股市有风险，投资需谨慎。

## License

MIT
