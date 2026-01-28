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
