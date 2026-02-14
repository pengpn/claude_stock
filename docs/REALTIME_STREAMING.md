# 实时流式输出实现

## 功能说明

实现了**真正的实时流式输出**，让用户在AI生成内容的同时就能看到内容逐步显示，而不是等待整个步骤完成。

## 用户体验

### 优化前
```
用户等待... (无响应)
    ↓
10秒后，综合分析完整内容一次性显示
    ↓
继续等待... (无响应)
    ↓
38秒后，多头观点完整内容一次性显示
```

### 优化后
```
用户等待...
    ↓
3秒后，看到"正在生成..."提示
    ↓
内容逐字符/逐句显示：
"根据财务数据分析..."  ← 即时显示
"该公司ROE为..."      ← 继续追加
"从估值角度看..."      ← 继续追加
▋ (光标闪烁)          ← 表示正在生成
```

**关键改进：**
- ✨ 立即反馈：用户不再等待，看到实时生成
- 💫 视觉反馈："正在生成..."提示 + 闪烁光标
- 📜 自动滚动：新步骤出现时自动滚动到可见区域
- 🎯 状态清晰：明确知道哪个步骤正在生成

## 实现架构

### 数据流

```
GLM API (流式响应)
    ↓
  每个delta立即返回
    ↓
Go Backend callback
    ↓
  立即发送SSE事件
    ↓
Gin Flush() 立即推送
    ↓
前端SSE客户端接收
    ↓
  解析delta内容
    ↓
Vue响应式更新DOM
    ↓
  用户看到新内容
```

**关键：每一步都是立即的，没有批处理或缓冲**

## 后端实现

### 1. GLM客户端流式处理

`backend/go-api/internal/llm/glm.go`：

```go
// 逐行读取SSE响应
reader := bufio.NewReader(resp.Body)
for {
    line, err := reader.ReadBytes('\n')
    // ... 解析SSE格式

    // 提取delta内容
    content := streamResp.Choices[0].Delta.Content
    if content != "" {
        // 立即调用callback，不批处理
        if err := callback(content); err != nil {
            return err
        }
    }
}
```

**关键点：**
- 使用 `bufio.Reader` 逐行读取
- 每收到一个delta立即callback
- 不累积，不批处理

### 2. Orchestrator立即推送

`backend/go-api/internal/service/orchestrator.go`：

```go
callback := func(delta string) error {
    content += delta

    // 立即发送到eventChan，不等待
    eventChan <- SSEEvent{
        Event: "analysis_step",
        Data: map[string]interface{}{
            "step":     string(step),
            "role":     stepName,
            "content":  delta,  // 只发送delta，不发送累积内容
            "progress": progress,
        },
    }
    return nil
}
```

**关键点：**
- 每个delta立即创建SSE事件
- channel缓冲区防止阻塞
- 发送delta而非全量内容（减少数据传输）

### 3. Handler立即Flush

`backend/go-api/internal/handler/analyze.go`：

```go
c.Stream(func(w io.Writer) bool {
    event, ok := <-eventChan
    if !ok {
        return false
    }

    // 发送SSE事件
    fmt.Fprintf(w, "event: %s\n", event.Event)
    fmt.Fprintf(w, "data: %s\n\n", dataJSON)

    // 立即flush，确保数据推送到客户端
    c.Writer.Flush()

    return true
})
```

**关键点：**
- 每个事件立即Flush
- 不等待缓冲区满
- SSE格式正确（event + data + 空行）

## 前端实现

### 1. 数据结构

`frontend/miniapp/src/pages/index/index.vue`：

```javascript
{
  step: 'comprehensive',     // 步骤标识
  role: '综合分析',          // 显示名称
  content: '累积的内容...',  // 累积显示的内容
  expanded: true,            // 展开状态
  streaming: true            // 正在流式生成中
}
```

**关键：`streaming` 状态控制视觉反馈**

### 2. 流式追加内容

```javascript
sse.addEventListener('analysis_step', (e) => {
    const data = e.data  // delta内容

    const existingIndex = results.value.findIndex(r => r.step === data.step)

    if (existingIndex >= 0) {
        // 追加delta到已有内容
        results.value[existingIndex].content += data.content
        results.value[existingIndex].streaming = true
    } else {
        // 首次出现，创建新步骤
        results.value.push({
            step: data.step,
            role: data.role,
            content: data.content,
            expanded: true,
            streaming: true
        })

        // 自动滚动到新步骤
        setTimeout(() => {
            uni.pageScrollTo({
                scrollTop: res[0].top,
                duration: 300
            })
        }, 100)
    }
})
```

**关键点：**
- 使用 `+=` 追加内容（不是替换）
- Vue响应式自动更新DOM
- 新步骤出现时自动滚动

### 3. 视觉反馈

```vue
<view class="card-header">
  <text class="card-icon">📊</text>
  <text class="card-title">{{ result.role }}</text>
  <view class="header-right">
    <!-- 流式生成提示 -->
    <text v-if="result.streaming" class="typing-indicator">
      正在生成...
    </text>
    <text class="expand-icon">▼</text>
  </view>
</view>

<view class="card-content">
  <text class="content-text">{{ result.content }}</text>
  <!-- 闪烁光标 -->
  <view v-if="result.streaming" class="cursor-blink">▋</view>
</view>
```

**视觉元素：**
1. **"正在生成..."** - 蓝色文字，脉冲动画
2. **闪烁光标 ▋** - 模拟打字机效果
3. **自动滚动** - 新内容出现时滚动到可见区域

### 4. CSS动画

```css
/* 脉冲动画：正在生成提示 */
.typing-indicator {
  color: #1890ff;
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

/* 闪烁动画：光标 */
.cursor-blink {
  color: #1890ff;
  animation: blink 1s step-end infinite;
}

@keyframes blink {
  0%, 50% { opacity: 1; }
  51%, 100% { opacity: 0; }
}
```

## SSE客户端优化

### 缓冲区处理

`frontend/miniapp/src/utils/sse.js`：

```javascript
parseSSE(text) {
    // 将新数据追加到缓冲区
    this.buffer += text

    // 分割行，保留不完整的行
    const lines = this.buffer.split('\n')
    if (!this.buffer.endsWith('\n')) {
        this.buffer = lines.pop() || ''
    } else {
        this.buffer = ''
    }

    // 逐行解析
    lines.forEach(line => {
        if (line.startsWith('event:')) {
            event = line.substring(6).trim()
        } else if (line.startsWith('data:')) {
            data = line.substring(5).trim()
        } else if (line === '') {
            // 消息完成，立即触发
            const parsedData = JSON.parse(data)
            this.emit(event, parsedData)
        }
    })
}
```

**关键：**
- 处理跨chunk的数据分包
- 保留不完整的行到下次
- 完整消息立即emit

## 性能优化

### 1. 减少不必要的渲染

```javascript
// ❌ 错误：每次都替换整个content
results.value[index].content = fullContent

// ✅ 正确：只追加delta
results.value[index].content += delta
```

### 2. 使用v-show而非v-if

```vue
<!-- ✅ v-show: DOM保留，只是隐藏 -->
<view v-show="result.expanded" class="card-content">
  {{ result.content }}
</view>

<!-- ❌ v-if: 频繁创建/销毁DOM -->
<view v-if="result.expanded" class="card-content">
  {{ result.content }}
</view>
```

### 3. 限流（如果需要）

```javascript
let lastUpdate = Date.now()
const THROTTLE_MS = 50  // 50ms内最多更新一次

callback := func(delta string) error {
    content += delta

    // 限流：避免更新太频繁
    now := Date.now()
    if (now - lastUpdate < THROTTLE_MS) {
        return nil
    }
    lastUpdate = now

    // 发送更新
    eventChan <- SSEEvent{...}
}
```

**注意：当前实现没有限流，因为GLM API已经有合理的频率**

## 调试方法

### 1. 前端控制台

```javascript
console.log('收到分析步骤:', data.step, '内容片段:', data.content)
```

查看每次收到的delta大小和频率

### 2. 后端日志

```go
log.Printf("发送delta: step=%s, len=%d", step, len(delta))
```

### 3. 网络面板

在微信开发者工具的网络面板中：
- 查看SSE连接状态
- 观察数据流是否持续
- 检查是否有延迟

## 常见问题

### Q1: 看不到流式效果，内容一次性显示

**可能原因：**
1. 网络代理或缓存导致批处理
2. GLM API本身批量返回
3. 前端缓冲问题

**解决：**
- 检查网络环境
- 查看调试日志确认delta频率
- 尝试不同的GLM模型

### Q2: 光标和"正在生成"不消失

**原因：** streaming状态没有正确清除

**解决：**
```javascript
sse.addEventListener('done', () => {
    // 清除所有streaming标记
    results.value.forEach(r => r.streaming = false)
})
```

### Q3: 内容显示卡顿

**原因：** DOM更新太频繁

**解决：** 添加限流（见性能优化部分）

### Q4: 多个步骤同时流式时混乱

**原因：** 并行执行导致事件交错

**现状：** 已正确处理，通过 `step` 字段区分

## 测试建议

### 功能测试

1. **基础流式**
   - [ ] 内容逐步显示
   - [ ] "正在生成..."出现
   - [ ] 光标闪烁
   - [ ] 完成后标识消失

2. **并行流式**
   - [ ] 多头和空头同时流式
   - [ ] 内容不混淆
   - [ ] 各自独立显示

3. **交互测试**
   - [ ] 流式时可以折叠/展开
   - [ ] 滚动不影响内容追加
   - [ ] 新步骤自动滚动到可见区域

### 性能测试

1. **长内容**
   - 测试1000+字的生成
   - 观察是否卡顿

2. **快速生成**
   - 使用glm-4-flash
   - 检查是否跟得上

3. **网络波动**
   - 模拟慢速网络
   - 检查是否正常accumulate

## 未来优化

### 1. 打字机效果增强

```javascript
// 按字符逐个显示，而不是一次性追加
let displayQueue = []
let isDisplaying = false

function enqueueContent(delta) {
    displayQueue.push(...delta.split(''))
    if (!isDisplaying) {
        displayNextChar()
    }
}

function displayNextChar() {
    if (displayQueue.length === 0) {
        isDisplaying = false
        return
    }

    isDisplaying = true
    const char = displayQueue.shift()
    results.value[index].content += char

    setTimeout(displayNextChar, 50)  // 50ms per char
}
```

### 2. 语音播报

```javascript
if (settings.voiceEnabled) {
    const utterance = new SpeechSynthesisUtterance(delta)
    speechSynthesis.speak(utterance)
}
```

### 3. 实时翻译

```javascript
// 边生成边翻译
sse.addEventListener('analysis_step', async (e) => {
    const translated = await translate(e.data.content)
    // 同时显示原文和译文
})
```

## 修改的文件

### 后端

1. **`backend/go-api/internal/service/orchestrator.go`**
   - 添加 `time` 导入
   - 优化callback立即发送
   - 添加流式标记

2. **`backend/go-api/internal/llm/glm.go`**
   - 保持逐行读取
   - 立即callback（无变化，已经是最优）

3. **`backend/go-api/internal/handler/analyze.go`**
   - 保持Flush（无变化，已经正确）

### 前端

1. **`frontend/miniapp/src/pages/index/index.vue`**
   - 添加 `streaming` 状态
   - 添加 `currentStreamingStep` ref
   - 视觉反馈组件
   - 自动滚动逻辑
   - CSS动画

2. **`frontend/miniapp/src/utils/sse.js`**
   - 缓冲区处理（已在之前修复）

## 性能指标

**目标：**
- 首字延迟 < 500ms（用户看到第一个字符）
- 字符延迟 < 100ms（相邻字符出现间隔）
- 完整步骤 < 30s（单个步骤完成时间）

**监控：**
```javascript
let firstCharTime = null
sse.addEventListener('analysis_step', (e) => {
    const now = Date.now()
    if (!firstCharTime) {
        firstCharTime = now
        console.log('首字延迟:', now - requestStartTime)
    } else {
        console.log('字符间隔:', now - lastCharTime)
    }
    lastCharTime = now
})
```

## 日期

实现日期：2026-01-28
状态：✅ 已完成并测试
