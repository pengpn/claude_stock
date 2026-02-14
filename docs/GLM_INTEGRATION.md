# GLM (智谱AI) 集成说明

## 概述

项目现已支持两种 LLM 提供商：
- **Claude** (Anthropic)
- **GLM** (智谱AI) - 国产大模型

## 配置 GLM

### 1. 获取 API 密钥

访问 [智谱AI开放平台](https://open.bigmodel.cn/) 注册并获取 API 密钥。

### 2. 配置环境变量

编辑 `.env` 文件：

```bash
# 选择 LLM 提供商
LLM_PROVIDER=glm

# GLM 配置
GLM_API_KEY=your-actual-api-key-here
GLM_BASE_URL=https://open.bigmodel.cn/api/paas/v4
GLM_MODEL=glm-4-plus
```

### 3. 可用模型

- `glm-4-plus` (推荐，默认)
- `glm-4-flash`
- `glm-4-air`
- 其他模型请参考智谱AI文档

## 切换 LLM

只需修改 `.env` 中的 `LLM_PROVIDER`：

```bash
# 使用 GLM
LLM_PROVIDER=glm

# 或使用 Claude
LLM_PROVIDER=claude
```

重启 Go API 服务后生效。

## API 兼容性

GLM 客户端实现了与 Claude 相同的 `LLMClient` 接口，支持：
- ✅ 流式响应 (SSE)
- ✅ 5 个分析角色（综合分析、多头、空头、交易员、风险委员会）
- ✅ 中文提示词优化
- ✅ 错误处理和降级

## 测试

```bash
# 重启服务
./scripts/start-dev.sh

# 测试分析接口
curl -X POST http://localhost:8000/api/v1/analyze \
  -H "Content-Type: application/json" \
  -d '{"code":"000858"}' \
  --no-buffer
```

## 故障排查

### GLM_API_KEY 未配置

```
2026/01/28 15:44:03 GLM_API_KEY未配置
```

**解决**: 在 `.env` 中设置有效的 GLM API 密钥

### API 403 错误

```
GLM API错误 (状态码 403): ...
```

**可能原因**:
1. API 密钥无效或过期
2. 账户余额不足
3. API 密钥权限不足

**解决**: 检查智谱AI控制台的密钥状态和余额

### 模型不存在

```
GLM API错误: model not found
```

**解决**: 检查 `GLM_MODEL` 配置，确保使用支持的模型名称

## 性能对比

| 特性 | GLM-4-Plus | Claude Sonnet 4.5 |
|------|------------|-------------------|
| 响应速度 | 快 | 中等 |
| 中文理解 | 优秀 | 优秀 |
| 成本 | 较低 | 较高 |
| 可用性 | 国内直连 | 需要代理 |

## 限制

1. GLM API 有速率限制，请参考智谱AI文档
2. 不同模型有不同的 token 限制
3. 流式响应可能在网络不稳定时中断

## 支持

- 智谱AI文档: https://open.bigmodel.cn/dev/api
- 智谱AI控制台: https://open.bigmodel.cn/console
