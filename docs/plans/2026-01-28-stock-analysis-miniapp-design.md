# 智能股票投资分析微信小程序 - 设计文档

**项目名称**: 智能股票投资分析小程序
**设计日期**: 2026-01-28
**版本**: v1.0 MVP

## 一、项目概述

### 1.1 项目目标
开发一个微信小程序，用户输入股票代码/名称后，通过AI多角色分析提供投资决策参考。

### 1.2 核心功能
- 用户输入股票代码或名称
- 5步AI分析流程：
  1. 综合分析（资深分析师视角）
  2. 多空辩论 - 多头观点
  3. 多空辩论 - 空头观点
  4. 交易员决策
  5. 风险委员会 + 最终决策
- 实时流式展示分析过程
- 给出明确的投资建议（买入/持有/卖出）和信心指数

### 1.3 MVP范围
- ✅ 仅提供股票分析功能
- ✅ 无需用户登录系统
- ✅ 不保存历史记录
- ✅ 本地开发调试（无需服务器、域名、备案）

---

## 二、技术架构

### 2.1 技术栈选择

| 层级 | 技术选型 | 说明 |
|------|---------|------|
| 前端 | Vue 3 + uni-app + TypeScript | 开发微信小程序 |
| API服务 | Go + Gin | 高性能API网关，处理请求路由和流式响应 |
| 数据分析 | Python + Flask/FastAPI | 股票数据获取和财务分析 |
| 数据源 | akshare | 免费A股数据接口 |
| AI引擎 | Claude API (Anthropic) | LLM角色扮演和分析生成 |
| 部署方式 | 本地运行 | 开发阶段仅在本地调试 |

### 2.2 系统架构图

```
┌─────────────────────────────────────┐
│   微信开发者工具 (小程序前端)        │
│   Vue 3 + uni-app                   │
└───────────────┬─────────────────────┘
                │ HTTP: localhost:8080
                ↓
┌─────────────────────────────────────┐
│   Go API 服务 (Gin)                 │
│   - 请求路由和参数校验               │
│   - SSE 流式响应                    │
│   - Claude API 调用管理             │
│   - LLM 抽象层（支持模型切换）       │
└───────┬─────────────────────────────┘
        │ HTTP: localhost:5000
        ↓
┌─────────────────────────────────────┐
│   Python 分析服务 (Flask/FastAPI)   │
│   - akshare 数据获取                │
│   - 财务指标计算                    │
│   - 估值分析                        │
│   - 风险检测                        │
└───────┬─────────────────────────────┘
        │
        ↓
┌─────────────────────────────────────┐
│   Claude API (Anthropic)            │
│   - AI 角色扮演                     │
│   - 流式文本生成                    │
└─────────────────────────────────────┘
```

---

## 三、核心模块设计

### 3.1 Go API 服务

#### 3.1.1 主要职责
- 接收小程序HTTP请求
- 验证股票代码有效性
- 调用Python服务获取结构化数据
- 编排5步AI分析流程
- 通过SSE流式返回分析结果
- 管理Claude API调用

#### 3.1.2 核心接口

**分析接口**
```http
POST /api/v1/analyze
Content-Type: application/json

请求体:
{
  "code": "000630",
  "name": "铜陵有色"  // 可选
}

响应: Server-Sent Events (SSE)

事件流:
event: progress
data: {"step": "fetching_data", "message": "正在获取股票数据...", "progress": 10}

event: analysis_step
data: {"step": "comprehensive", "role": "分析师", "content": "铜陵有色是...", "progress": 20}

event: analysis_step
data: {"step": "debate_bull", "role": "多头", "content": "看多理由：1...", "progress": 40}

event: analysis_step
data: {"step": "debate_bear", "role": "空头", "content": "看空风险：1...", "progress": 60}

event: analysis_step
data: {"step": "trader", "role": "交易员", "content": "操作建议...", "progress": 80}

event: final_decision
data: {
  "step": "final",
  "role": "风险委员会",
  "recommendation": "hold",  // buy/hold/sell
  "confidence": 75,          // 0-100
  "summary": "综合建议...",
  "content": "完整分析...",
  "progress": 100
}

event: done
data: {"message": "分析完成"}
```

#### 3.1.3 目录结构
```
go-api/
├── cmd/
│   └── main.go                 # 入口文件
├── internal/
│   ├── handler/
│   │   └── analyze.go          # HTTP handler
│   ├── service/
│   │   ├── orchestrator.go     # 分析流程编排
│   │   └── stock_service.go    # 业务逻辑
│   ├── llm/
│   │   ├── client.go           # LLM接口定义
│   │   ├── claude.go           # Claude实现
│   │   └── prompt.go           # Prompt模板
│   ├── client/
│   │   └── python_client.go    # Python服务客户端
│   └── model/
│       └── types.go            # 数据结构
├── config/
│   └── config.go               # 配置管理
├── go.mod
└── go.sum
```

### 3.2 Python 分析服务

#### 3.2.1 主要职责
- 使用akshare获取股票实时行情和财务数据
- 计算关键财务指标（ROE、PE、PB、资产负债率等）
- 执行估值分析（DCF、DDM、相对估值）
- 财务异常检测
- 行业对比分析

#### 3.2.2 核心接口

**数据分析接口**
```http
POST /analyze
Content-Type: application/json

请求体:
{
  "code": "000630"
}

响应:
{
  "code": "000630",
  "name": "铜陵有色",
  "basic_info": {
    "industry": "有色金属-铜冶炼",
    "market_cap": 580.5,        // 亿元
    "pe_ttm": 12.5,
    "pb": 1.2,
    "latest_price": 7.28,
    "price_change_pct": -1.35
  },
  "financial_metrics": {
    "roe": 15.2,                // %
    "roa": 6.8,
    "gross_margin": 8.5,
    "net_margin": 4.2,
    "debt_ratio": 55.3,
    "current_ratio": 1.45,
    "revenue_growth": 8.5,      // 同比增长%
    "profit_growth": 12.3
  },
  "valuation": {
    "dcf_value": 8.50,
    "current_price": 7.28,
    "upside": 16.8,             // %
    "undervalued": true,
    "pb_percentile": 35.2,      // 历史分位数
    "pe_percentile": 28.5
  },
  "risks": [
    "铜价波动风险",
    "资产负债率偏高",
    "应收账款增长较快"
  ],
  "industry_comparison": {
    "rank": 3,
    "total": 10,
    "roe_rank": 2,
    "peers": ["江西铜业", "云南铜业", "紫金矿业"]
  },
  "analysis_date": "2026-01-28"
}
```

#### 3.2.3 目录结构
```
python-analysis/
├── app.py                      # Flask/FastAPI入口
├── services/
│   ├── data_fetcher.py         # 数据获取（复用china-stock-analysis）
│   ├── financial_analyzer.py   # 财务分析
│   └── valuation_calculator.py # 估值计算
├── utils/
│   ├── cache.py                # 缓存工具
│   └── logger.py               # 日志
├── requirements.txt
└── config.py
```

### 3.3 前端小程序

#### 3.3.1 页面结构

**主页面 (pages/index/index.vue)**

```
┌─────────────────────────────────┐
│      智能股票分析助手            │
├─────────────────────────────────┤
│  🔍 [输入框: 股票代码/名称]      │
│     [开始分析] 按钮              │
├─────────────────────────────────┤
│  进度条: ████████░░░ 80%        │
├─────────────────────────────────┤
│  📊 综合分析                     │
│  ┌───────────────────────────┐  │
│  │ [分析内容 - 流式展示]      │  │
│  │ 支持展开/收起              │  │
│  └───────────────────────────┘  │
├─────────────────────────────────┤
│  🐂 多头观点                     │
│  ┌───────────────────────────┐  │
│  │ [看多理由...]             │  │
│  └───────────────────────────┘  │
├─────────────────────────────────┤
│  🐻 空头观点                     │
│  ┌───────────────────────────┐  │
│  │ [看空风险...]             │  │
│  └───────────────────────────┘  │
├─────────────────────────────────┤
│  💼 交易员决策                   │
│  ┌───────────────────────────┐  │
│  │ [操作建议...]             │  │
│  └───────────────────────────┘  │
├─────────────────────────────────┤
│  ✅ 最终决策                     │
│  ┌───────────────────────────┐  │
│  │ 推荐操作: 【持有】         │  │
│  │ 信心指数: 75/100 ⭐⭐⭐⭐   │  │
│  │ [综合建议...]             │  │
│  └───────────────────────────┘  │
└─────────────────────────────────┘
```

#### 3.3.2 关键功能

**SSE 事件监听**
```javascript
const eventSource = new EventSource('/api/v1/analyze')

eventSource.addEventListener('analysis_step', (e) => {
  const data = JSON.parse(e.data)
  // 流式更新对应步骤的内容
  updateStep(data.step, data.content, data.progress)
})

eventSource.addEventListener('final_decision', (e) => {
  const data = JSON.parse(e.data)
  // 显示最终决策
  showFinalDecision(data)
})
```

**交互特性**
- 加载动画：骨架屏 + 进度条
- 流式打字效果：模拟AI实时思考
- 内容展开/收起：长文本优化阅读
- Markdown渲染：支持粗体、列表、表格
- 错误处理：友好提示 + 重试

#### 3.3.3 目录结构
```
miniapp/
├── pages/
│   └── index/
│       ├── index.vue           # 主页面
│       └── components/
│           ├── SearchBar.vue   # 搜索栏
│           ├── AnalysisCard.vue # 分析卡片
│           └── FinalDecision.vue # 最终决策
├── api/
│   └── stock.js                # API调用封装
├── utils/
│   ├── sse.js                  # SSE工具
│   └── markdown.js             # Markdown渲染
├── static/
├── App.vue
├── main.js
├── manifest.json
└── pages.json
```

---

## 四、AI 角色设计

### 4.1 分析流程

```
用户输入股票代码
    ↓
Python获取数据 (5-10秒)
    ↓
步骤1: 综合分析 (5-8秒)
    ↓
步骤2: 多头观点 (5-8秒)
    ↓
步骤3: 空头观点 (5-8秒)
    ↓
步骤4: 交易员决策 (5-8秒)
    ↓
步骤5: 最终决策 (5-8秒)
    ↓
完成 (总计30-45秒)
```

### 4.2 五个AI角色详细设计

#### 角色1: 综合分析师 (Comprehensive Analyst)

**角色定位**: 资深投资分析师，客观中立

**输入数据**:
- Python返回的所有结构化数据
- 基本信息、财务指标、估值数据、风险信号

**输出内容** (200-300字):
- 公司基本情况（行业地位、主营业务）
- 财务健康度评估（ROE、资产负债率、现金流）
- 估值水平（PE/PB历史分位数）
- 关键风险提示

**Prompt模板**:
```
你是一位资深的A股投资分析师，擅长客观中立地分析上市公司。

请基于以下数据，对【{stock_name}({stock_code})】进行综合分析：

【基本信息】
- 行业: {industry}
- 市值: {market_cap}亿元
- 最新价: {price}元
- PE: {pe}, PB: {pb}

【财务指标】
- ROE: {roe}%
- 资产负债率: {debt_ratio}%
- 营收增长: {revenue_growth}%
- 净利润增长: {profit_growth}%

【估值分析】
- DCF估值: {dcf_value}元
- 当前价格: {current_price}元
- 上涨空间: {upside}%

【风险信号】
{risks}

请从以下角度进行分析：
1. 公司基本情况和行业地位
2. 财务健康度（盈利能力、偿债能力、成长性）
3. 估值水平（是否被低估/高估）
4. 主要风险点

要求：
- 客观中立，基于数据
- 200-300字
- 结构清晰，分点阐述
```

#### 角色2: 多头投资者 (Bull Investor)

**角色定位**: 乐观的多头，寻找投资亮点

**输入数据**:
- 综合分析结果
- 结构化财务数据

**输出内容** (150-200字):
- 3-5个看多的核心理由
- 强调估值优势、成长潜力、行业前景
- 积极正面的投资逻辑

**Prompt模板**:
```
你是一位乐观的多头投资者，擅长挖掘股票的投资价值和上涨潜力。

基于以下综合分析和数据，请给出【{stock_name}】的看多观点：

【综合分析】
{comprehensive_analysis}

【关键数据】
- ROE: {roe}%（行业排名: {roe_rank}/{total}）
- 估值上涨空间: {upside}%
- 营收增长: {revenue_growth}%

请从多头角度回答：
1. 这只股票最吸引人的3-5个投资亮点是什么？
2. 为什么现在是好的买入时机？
3. 未来的上涨驱动力在哪里？

要求：
- 积极正面，但基于数据事实
- 150-200字
- 突出投资价值
```

#### 角色3: 空头投资者 (Bear Investor)

**角色定位**: 谨慎的空头，挑剔风险

**输入数据**:
- 综合分析结果
- 财务风险信号

**输出内容** (150-200字):
- 3-5个看空的核心风险
- 质疑估值、财务风险、行业逆风
- 批判性的投资视角

**Prompt模板**:
```
你是一位谨慎的空头投资者，擅长识别风险和质疑过度乐观的预期。

基于以下信息，请给出【{stock_name}】的看空观点：

【综合分析】
{comprehensive_analysis}

【风险信号】
{risks}

【关键数据】
- 资产负债率: {debt_ratio}%
- 当前PE: {pe}（历史分位数: {pe_percentile}%）
- 行业: {industry}

请从空头角度回答：
1. 这只股票最大的3-5个风险点是什么？
2. 为什么当前估值可能不便宜？
3. 哪些因素可能导致下跌？

要求：
- 批判谨慎，但基于逻辑
- 150-200字
- 突出风险因素
```

#### 角色4: 实战交易员 (Trader)

**角色定位**: 经验丰富的交易员，给出具体操作

**输入数据**:
- 综合分析
- 多空观点
- 市场数据

**输出内容** (150-200字):
- 明确操作建议（买入/持有/卖出）
- 建议仓位（轻仓/中仓/重仓）
- 参考买入价、止损价
- 持有周期建议

**Prompt模板**:
```
你是一位实战经验丰富的A股交易员，擅长将分析转化为具体的交易决策。

基于以下多空观点，给出【{stock_name}】的交易建议：

【综合分析】
{comprehensive_analysis}

【多头观点】
{bull_case}

【空头观点】
{bear_case}

【当前价格】{current_price}元

请给出具体的交易建议：
1. 操作方向（买入/持有/卖出）
2. 建议仓位（轻仓5-10% / 中仓10-20% / 重仓20%+）
3. 参考买入价位区间
4. 止损位设置
5. 预期持有周期

要求：
- 具体可执行
- 风险收益比考量
- 150-200字
```

#### 角色5: 风险委员会 + 最终决策 (Risk Committee & Final Decision)

**角色定位**: 风险管理委员会，综合决策

**输入数据**:
- 前4步所有分析结果

**输出内容** (200-250字):
- 风险评估（高/中/低）
- 综合投资建议（买入/持有/卖出）
- 信心指数（0-100）
- 最终结论摘要

**Prompt模板**:
```
你是投资决策委员会的风险管理官，负责综合各方意见给出最终决策。

请基于以下完整分析链，给出【{stock_name}】的最终投资建议：

【综合分析】
{comprehensive_analysis}

【多头观点】
{bull_case}

【空头观点】
{bear_case}

【交易员建议】
{trader_decision}

请给出最终决策：
1. 风险等级评估（高/中/低风险）
2. 综合投资建议（买入/持有/卖出）
3. 信心指数（0-100，表示建议的确定性）
4. 决策理由总结（综合权衡多空观点）

JSON格式输出：
{
  "recommendation": "buy/hold/sell",
  "risk_level": "high/medium/low",
  "confidence": 75,
  "summary": "最终决策理由（200字内）"
}

要求：
- 平衡风险和收益
- 给出明确结论
- 200-250字
```

### 4.3 LLM 调用参数

| 参数 | 值 | 说明 |
|------|---|------|
| model | claude-3-5-sonnet-20241022 | 平衡性能和成本 |
| max_tokens | 800 | 每个角色输出控制在300字内 |
| temperature | 0.7 | 保持创造性但不过度发散 |
| stream | true | 流式响应 |
| system | 各角色的system prompt | 定义角色身份 |

---

## 五、数据流设计

### 5.1 完整请求流程

```
1. 用户在小程序输入 "000630"
   ↓
2. 前端发送POST请求到 Go API
   POST http://localhost:8080/api/v1/analyze
   Body: {"code": "000630"}
   ↓
3. Go验证股票代码格式
   ↓
4. Go调用Python服务
   POST http://localhost:5000/analyze
   ↓
5. Python执行：
   - akshare获取实时数据
   - 计算财务指标
   - 估值分析
   - 风险检测
   返回结构化JSON
   ↓
6. Go接收数据后，开始5步AI分析
   ↓
7. 步骤1: 调用Claude API（综合分析）
   - 构造prompt（包含Python数据）
   - 流式返回分析内容
   - 通过SSE推送给前端
   ↓
8. 步骤2: 调用Claude API（多头观点）
   - 基于步骤1结果 + 数据
   - 流式返回
   ↓
9. 步骤3: 调用Claude API（空头观点）
   - 基于步骤1结果 + 数据
   - 流式返回
   ↓
10. 步骤4: 调用Claude API（交易员）
    - 基于步骤1-3结果
    - 流式返回
    ↓
11. 步骤5: 调用Claude API（最终决策）
    - 基于所有前序分析
    - 返回JSON格式决策
    - 流式返回
    ↓
12. 前端实时渲染每个步骤
    - 显示进度条
    - 流式打字效果
    - 最终高亮显示决策
```

### 5.2 SSE事件定义

| Event | Data | 说明 |
|-------|------|------|
| progress | `{"step": "fetching_data", "message": "...", "progress": 10}` | 进度更新 |
| analysis_step | `{"step": "comprehensive", "role": "分析师", "content": "...", "progress": 20}` | 分析步骤 |
| final_decision | `{"recommendation": "hold", "confidence": 75, "summary": "..."}` | 最终决策 |
| error | `{"code": "STOCK_NOT_FOUND", "message": "..."}` | 错误信息 |
| done | `{"message": "分析完成"}` | 完成标记 |

---

## 六、本地开发部署

### 6.1 开发环境要求

| 组件 | 版本要求 |
|------|---------|
| Go | >= 1.20 |
| Python | >= 3.10 |
| Node.js | >= 16 |
| 微信开发者工具 | 最新稳定版 |

### 6.2 启动步骤

#### Step 1: 配置环境变量

创建 `.env` 文件：
```bash
# Claude API配置
CLAUDE_API_KEY=sk-ant-xxx

# Python服务地址
PYTHON_SERVICE_URL=http://localhost:5000

# 服务端口
GO_API_PORT=8080
PYTHON_API_PORT=5000
```

#### Step 2: 启动Python分析服务

```bash
cd backend/python-analysis
pip install -r requirements.txt
python app.py

# 输出:
# * Running on http://localhost:5000
```

#### Step 3: 启动Go API服务

```bash
cd backend/go-api
go mod download
go run cmd/main.go

# 输出:
# Server running on :8080
```

#### Step 4: 启动小程序

```bash
cd frontend/miniapp
npm install
npm run dev:mp-weixin

# 然后在微信开发者工具中导入 dist/dev/mp-weixin
```

#### Step 5: 配置微信开发者工具

1. 打开微信开发者工具
2. 导入项目：选择 `frontend/miniapp/dist/dev/mp-weixin`
3. 设置 → 项目设置 → 勾选"不校验合法域名、web-view、TLS版本"
4. 开始调试

### 6.3 测试用例

**测试股票代码**:
- 000630 (铜陵有色) - 有色金属
- 600519 (贵州茅台) - 白酒
- 000858 (五粮液) - 白酒
- 601318 (中国平安) - 保险

**预期结果**:
- 数据获取时间: 5-10秒
- 单步分析时间: 5-8秒
- 总耗时: 30-45秒
- 流式展示顺畅，无明显卡顿

---

## 七、项目结构

```
stock-analysis-miniapp/
├── README.md
├── .env.example
├── .gitignore
├── docs/
│   └── plans/
│       └── 2026-01-28-stock-analysis-miniapp-design.md
├── backend/
│   ├── go-api/
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── internal/
│   │   │   ├── handler/
│   │   │   │   └── analyze.go
│   │   │   ├── service/
│   │   │   │   ├── orchestrator.go
│   │   │   │   └── stock_service.go
│   │   │   ├── llm/
│   │   │   │   ├── client.go
│   │   │   │   ├── claude.go
│   │   │   │   └── prompts.go
│   │   │   ├── client/
│   │   │   │   └── python_client.go
│   │   │   └── model/
│   │   │       └── types.go
│   │   ├── config/
│   │   │   └── config.go
│   │   ├── go.mod
│   │   └── go.sum
│   └── python-analysis/
│       ├── app.py
│       ├── services/
│       │   ├── data_fetcher.py
│       │   ├── financial_analyzer.py
│       │   └── valuation_calculator.py
│       ├── utils/
│       │   ├── cache.py
│       │   └── logger.py
│       ├── config.py
│       └── requirements.txt
└── frontend/
    └── miniapp/
        ├── pages/
        │   └── index/
        │       ├── index.vue
        │       └── components/
        │           ├── SearchBar.vue
        │           ├── AnalysisCard.vue
        │           └── FinalDecision.vue
        ├── api/
        │   └── stock.js
        ├── utils/
        │   ├── sse.js
        │   └── markdown.js
        ├── static/
        ├── App.vue
        ├── main.js
        ├── manifest.json
        ├── pages.json
        └── package.json
```

---

## 八、成本估算（本地开发）

### 8.1 基础设施成本
- ✅ 服务器: **免费**（本地运行）
- ✅ 域名: **免费**（localhost）
- ✅ SSL证书: **不需要**

### 8.2 API调用成本

**Claude API (Sonnet 3.5)**:
- 输入: $3 / 百万tokens
- 输出: $15 / 百万tokens

**单次分析估算**:
- 输入约 3,000 tokens (结构化数据 × 5次调用)
- 输出约 2,000 tokens (5个角色各400字)
- 单次成本: (3000×$3 + 2000×$15) / 1,000,000 ≈ **$0.04**

**月度估算**:
- 100次分析/月: $4
- 500次分析/月: $20
- 1000次分析/月: $40

**总计**: 开发阶段每月 **$10-20** (测试100-500次)

---

## 九、风险和限制

### 9.1 技术风险

| 风险 | 影响 | 应对措施 |
|------|------|---------|
| akshare数据不稳定 | 获取数据失败 | 降级策略：返回基本数据，提示部分功能不可用 |
| Claude API限流 | 分析失败 | 重试机制 + 错误提示 |
| SSE连接断开 | 分析中断 | 超时重连 + 状态保存 |
| 小程序request超时 | 用户等待时间过长 | 设置60秒超时，超时提示 |

### 9.2 业务限制

- ⚠️ 仅支持A股市场
- ⚠️ 数据有延迟（非实时Level-2行情）
- ⚠️ AI分析仅供参考，不构成投资建议
- ⚠️ 本地开发无法多人访问

### 9.3 合规风险

- ⚠️ 需在显眼位置声明"不构成投资建议"
- ⚠️ 避免使用"推荐买入""必涨"等诱导性词汇
- ⚠️ 不保证分析结果的准确性

---

## 十、后续迭代计划

### 10.1 Phase 2: 功能增强

- [ ] 用户登录系统（微信授权）
- [ ] 历史分析记录
- [ ] 自选股收藏
- [ ] 分析结果缓存（Redis）

### 10.2 Phase 3: 生产部署

- [ ] 云服务器部署
- [ ] 域名和SSL证书
- [ ] CDN加速
- [ ] 监控和日志

### 10.3 Phase 4: 模型切换

- [ ] LLM抽象层完善
- [ ] 接入国产大模型（通义千问/文心一言）
- [ ] 成本优化

---

## 十一、免责声明

**本系统提供的所有分析和建议仅供参考，不构成任何投资建议。**

- 股市有风险，投资需谨慎
- AI分析可能存在偏差和错误
- 用户需独立判断，自负盈亏
- 历史数据不代表未来表现
- 开发者不对投资损失承担任何责任

---

**设计完成日期**: 2026-01-28
**下一步**: 开始实现代码
