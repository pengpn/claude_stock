# 分析速度优化

## 问题描述

用户反馈：每个分析步骤显示很慢，整个分析流程需要2-3分钟才能完成。

## 原因分析

### 执行时间分解（优化前）

从日志分析实际执行时间：

```
综合分析:     7秒   ━━━
多头观点:    28秒   ━━━━━━━━━━━━
空头观点:    70秒   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━
交易员决策:  28秒   ━━━━━━━━━━━━
最终决策:    22秒   ━━━━━━━━━
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
总计: 155秒 (约2.5分钟)
```

### 根本原因

1. **完全串行执行**
   - 5个步骤依次执行，后一个必须等前一个完成
   - 多头观点等待综合分析完成
   - 空头观点等待多头完成
   - ...依次等待

2. **LLM API延迟**
   - 每次调用GLM API需要等待完整生成
   - 模型生成需要时间（取决于内容长度）
   - 网络往返延迟

3. **步骤依赖设计**
   ```go
   综合分析 → 多头观点 → 空头观点 → 交易员 → 最终决策
   ```
   每个步骤的输出作为下一步的输入

## 优化方案

### ✅ 优化1：并行执行多空观点（已实现）

**关键洞察**：多头和空头观点不互相依赖，可以同时执行

#### 优化前（串行）
```
综合分析 (7s)
    ↓
多头观点 (28s)  ← 等待综合完成
    ↓
空头观点 (70s)  ← 等待多头完成
    ↓
交易员 (28s)
    ↓
最终决策 (22s)
━━━━━━━━━━━━━━━━
总计: 155秒
```

#### 优化后（并行）
```
综合分析 (7s)
    ↓
多头观点 (28s)  ┐
              ├─ 并行执行，只需70秒
空头观点 (70s)  ┘
    ↓
交易员 (28s)
    ↓
最终决策 (22s)
━━━━━━━━━━━━━━━━
总计: 127秒 (节省28秒)
```

#### 实现代码

```go
// 并行执行多头和空头
var wg sync.WaitGroup
errChan := make(chan error, 2)

wg.Add(1)
go func() {
    defer wg.Done()
    if err := ao.runStepConcurrent(ctx, llm.StepDebateBull, "多头观点",
        llmData, results, resultsMutex, eventChan, 40); err != nil {
        errChan <- err
    }
}()

wg.Add(1)
go func() {
    defer wg.Done()
    if err := ao.runStepConcurrent(ctx, llm.StepDebateBear, "空头观点",
        llmData, results, resultsMutex, eventChan, 60); err != nil {
        errChan <- err
    }
}()

wg.Wait()
```

**关键点：**
- 使用 `sync.WaitGroup` 等待两个goroutine完成
- 使用 `sync.Mutex` 保护共享的 `results` map
- 使用 `errChan` 收集错误

### ✅ 优化2：使用更快的模型（已配置）

将GLM模型从 `glm-4.7` 改为 `glm-4-flash`

#### 模型对比

| 模型 | 速度 | 质量 | 适用场景 |
|------|------|------|----------|
| glm-4-flash | ⚡⚡⚡ 最快 | ⭐⭐ 良好 | 快速分析、对话 |
| glm-4-air | ⚡⚡ 较快 | ⭐⭐⭐ 优秀 | 平衡场景 |
| glm-4-plus | ⚡ 较慢 | ⭐⭐⭐⭐ 最强 | 深度分析、创作 |

#### 配置方式

修改 `.env` 文件：
```bash
GLM_MODEL=glm-4-flash  # 使用最快模型
```

**预期效果**：每个步骤速度提升30-50%

### 💡 优化3：调整生成参数（可选）

修改 `internal/llm/glm.go`：

```go
reqBody := glmRequest{
    Model: g.model,
    Messages: messages,
    Temperature: 0.5,      // 降低温度（0.7 → 0.5）加快生成
    MaxTokens:   500,      // 减少最大token数（800 → 500）
    Stream:      true,
}
```

**权衡**：
- ✅ 速度更快
- ⚠️ 内容可能更简短
- ⚠️ 多样性略降低

### 💡 优化4：优化提示词（可选）

缩短系统提示词，减少输入token数：

```go
// 优化前
"你是一位专业的股票分析师，请根据以下财务数据进行详细分析。分析要全面、深入，考虑多个维度..."

// 优化后
"作为股票分析师，基于财务数据提供简洁分析，聚焦关键要点..."
```

**效果**：减少10-20%处理时间

### 💡 优化5：缓存Python数据（未实现）

对于相同股票，缓存Python分析结果：

```go
// 伪代码
if cached := cache.Get(stockCode); cached != nil {
    return cached
}

data := pythonClient.Analyze(code)
cache.Set(stockCode, data, 5*time.Minute)
```

**效果**：重复查询速度提升90%

## 性能对比

### 优化前后对比

| 阶段 | 优化前 | 优化后（并行） | 优化后（并行+快速模型） |
|------|--------|----------------|-------------------------|
| 综合分析 | 7s | 7s | 4s |
| 多头观点 | 28s | ┐ | ┐ |
| 空头观点 | 70s | ┘ 70s (并行) | ┘ 40s (并行) |
| 交易员 | 28s | 28s | 15s |
| 最终决策 | 22s | 22s | 12s |
| **总计** | **155s** | **127s** | **71s** |
| **节省** | - | **-28s (-18%)** | **-84s (-54%)** |

### 预期效果

- ✅ **并行执行**：节省约18%时间
- ✅ **快速模型**：再节省30-40%时间
- 🎯 **总体优化**：从2.5分钟缩短到约1分钟

## 实现细节

### 修改的文件

1. **`backend/go-api/internal/service/orchestrator.go`**
   - 添加 `sync` 包导入
   - 修改 `Analyze` 方法：多空观点并行执行
   - 新增 `runStepConcurrent` 方法：线程安全的步骤执行

2. **`.env`**
   - 修改 `GLM_MODEL=glm-4-flash`（使用最快模型）

### 线程安全

使用 `sync.Mutex` 保护共享资源：

```go
resultsMutex := &sync.Mutex{}

// 在goroutine中线程安全地写入
mutex.Lock()
results[string(step)] = content
mutex.Unlock()
```

### 错误处理

使用channel收集并行执行的错误：

```go
errChan := make(chan error, 2)

// goroutine发送错误
errChan <- err

// 主线程检查错误
for err := range errChan {
    if err != nil {
        return err
    }
}
```

## 用户体验改进

### 优化前
```
用户输入股票代码
    ↓
等待 10秒 ... (看到进度条，但没有内容)
    ↓
综合分析出现
    ↓
等待 28秒 ... (进度条缓慢前进)
    ↓
多头观点出现
    ↓
等待 70秒 ... (用户开始焦虑)
    ↓
空头观点出现
    ↓
... 继续等待
    ↓
(2.5分钟后) 所有结果显示完毕
```

### 优化后
```
用户输入股票代码
    ↓
等待 7秒
    ↓
综合分析出现
    ↓
等待 35秒 (多头和空头同时生成)
    ↓
多头 + 空头同时出现！
    ↓
等待 15秒
    ↓
交易员决策出现
    ↓
等待 12秒
    ↓
(1分钟后) 所有结果显示完毕
```

**改进**：
- ⚡ 总时间减半
- 👀 更快看到结果
- 😊 用户体验明显提升

## 进一步优化方向

### 1. 流式展示优化

当前是完整步骤流式展示，可以优化为：
- 显示"正在分析..."占位符
- 逐字符流式显示（已实现）
- 显示预计剩余时间

### 2. 智能预热

在用户输入时开始预加载：
```javascript
// 前端
input.addEventListener('input', (e) => {
    if (e.target.value.length >= 6) {
        // 预先获取股票基本信息
        prefetchBasicInfo(e.target.value)
    }
})
```

### 3. 分阶段展示

不等所有步骤完成，逐步展示：
- 阶段1：综合分析 + 多空观点 (显示完立即可用)
- 阶段2：交易建议 + 最终决策 (后台继续生成)

### 4. 降级策略

检测网络或API延迟，自动调整：
```go
if latency > threshold {
    // 自动切换到更快的模型
    // 或减少分析深度
}
```

### 5. 本地缓存

使用Redis缓存分析结果：
- 相同股票5分钟内直接返回
- 盘中数据实时更新，盘后数据可缓存更久

## 测试建议

### 性能测试

1. **测试并行执行**
   ```bash
   # 启动服务
   ./scripts/start-dev.sh

   # 分析股票，记录时间
   time curl -X POST http://localhost:8000/api/v1/analyze \
     -H "Content-Type: application/json" \
     -d '{"code":"000858"}'
   ```

2. **对比不同模型**
   ```bash
   # 测试 glm-4-flash
   GLM_MODEL=glm-4-flash go run cmd/main.go

   # 测试 glm-4-plus
   GLM_MODEL=glm-4-plus go run cmd/main.go
   ```

3. **压力测试**
   ```bash
   # 10个并发请求
   ab -n 10 -c 2 -p request.json -T application/json \
     http://localhost:8000/api/v1/analyze
   ```

### 功能测试

- [ ] 多头和空头同时显示
- [ ] 错误处理正常（一个失败不影响另一个）
- [ ] 流式内容正确展示
- [ ] 前端折叠/展开功能正常
- [ ] 日志显示并行执行标记

## 监控指标

添加性能监控：

```go
import "time"

start := time.Now()
defer func() {
    duration := time.Since(start)
    log.Printf("分析总耗时: %v", duration)
}()
```

**关键指标**：
- 总分析时间
- 各步骤耗时
- 并行执行重叠时间
- LLM API响应时间

## 日期

优化实施：2026-01-28
预期效果：总时间从155秒降至70秒（提升54%）
