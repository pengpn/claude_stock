# 智能股票分析小程序

基于 Claude AI 的智能股票分析工具，提供专业的财务分析和投资建议。

## 项目简介

本项目是一个集成了人工智能能力的股票分析系统，通过 Claude AI 提供深度的财务分析、价值评估和投资建议。系统采用微服务架构，包含 Python 数据分析服务、Go API 网关和 uni-app 小程序前端。

## 技术栈

### 后端
- **Python 分析服务**
  - FastAPI: Web 框架
  - akshare: 股票数据获取
  - pandas/numpy: 数据处理和分析
  - 财务指标计算和价值评估

- **Go API 网关**
  - Gin: Web 框架
  - Claude API: AI 分析能力
  - SSE: 流式响应支持
  - 服务编排和请求代理

### 前端
- **uni-app 小程序**
  - Vue.js: 前端框架
  - uni-ui: UI 组件库
  - 支持微信小程序、H5 等多端

## 项目结构

```
claude_stock/
├── backend/
│   ├── python-analysis/        # Python 分析服务
│   │   ├── services/           # 业务服务
│   │   └── utils/              # 工具函数
│   └── go-api/                 # Go API 网关
│       ├── cmd/                # 主程序入口
│       ├── internal/           # 内部实现
│       │   ├── handler/        # HTTP 处理器
│       │   ├── service/        # 业务服务
│       │   ├── llm/            # LLM 抽象层
│       │   ├── client/         # 外部客户端
│       │   └── model/          # 数据模型
│       └── config/             # 配置文件
└── frontend/
    └── miniapp/                # uni-app 小程序
        ├── pages/              # 页面
        ├── api/                # API 封装
        ├── utils/              # 工具函数
        └── static/             # 静态资源

## 本地开发指南

### 环境要求

- Python 3.9+
- Go 1.21+
- Node.js 16+
- HBuilderX (用于小程序开发)

### 快速开始

1. **克隆项目**
```bash
git clone <repository-url>
cd claude_stock
```

2. **配置环境变量**
```bash
cp .env.example .env
# 编辑 .env 文件，填入你的 Claude API Key
```

3. **启动 Python 分析服务**
```bash
cd backend/python-analysis
pip install -r requirements.txt
python main.py
```

4. **启动 Go API 网关**
```bash
cd backend/go-api
go mod download
go run cmd/main.go
```

5. **启动小程序**
```bash
cd frontend/miniapp
npm install
npm run dev:mp-weixin
```

### API 文档

- Python 分析服务: http://localhost:8001/docs
- Go API 网关: http://localhost:8000

## 功能特性

- 实时股票数据获取
- 财务指标分析
- 价值评估模型
- AI 驱动的投资建议
- 流式分析结果展示
- 多端小程序支持

## 开发计划

详细的开发计划和任务安排请查看 `docs/plans/` 目录。

## 许可证

MIT License
