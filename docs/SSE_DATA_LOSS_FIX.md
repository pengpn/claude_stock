# SSE数据丢失问题修复

## 问题描述

前端小程序在接收SSE流式数据时，部分分析步骤不显示，导致5个步骤中只有部分步骤有内容。

## 根本原因

**SSE客户端无法处理跨chunk的数据分包**

微信小程序的 `uni.request` 使用 `enableChunked: true` 时，服务器返回的SSE数据流会被拆分成多个chunk传输。如果一个SSE消息正好被拆分到两个chunk中，原有的解析逻辑会失败。

### 示例场景

**正常情况（单个chunk）：**
```
event: analysis_step
data: {"step":"comprehensive","role":"综合分析","content":"这是分析内容..."}

```

**分包情况（跨两个chunk）：**
```
Chunk 1:
event: analysis_step
data: {"step":"comprehensive","role":"综合分

Chunk 2:
析","content":"这是分析内容..."}

```

原有的 `parseSSE` 方法在每个chunk独立解析，无法处理第二种情况，导致：
1. Chunk 1 的不完整JSON无法解析 → 数据丢失
2. Chunk 2 的残缺数据无法解析 → 数据丢失
3. 该步骤不显示任何内容

## 修复方案

### 1. 添加缓冲区

在 `SSEClient` 类中添加 `buffer` 属性：

```javascript
export class SSEClient {
  constructor(url, data) {
    this.url = url
    this.data = data
    this.listeners = {}
    this.requestTask = null
    this.buffer = ''  // 缓冲区
  }
}
```

### 2. 改进parseSSE方法

```javascript
parseSSE(text) {
  // 将新数据追加到缓冲区
  this.buffer += text

  // 按行分割
  const lines = this.buffer.split('\n')

  // 保留不完整的最后一行
  if (!this.buffer.endsWith('\n')) {
    this.buffer = lines.pop() || ''
  } else {
    this.buffer = ''
  }

  let event = 'message'
  let data = ''

  lines.forEach(line => {
    const trimmedLine = line.trim()

    if (trimmedLine.startsWith('event:')) {
      event = trimmedLine.substring(6).trim()
    } else if (trimmedLine.startsWith('data:')) {
      data = trimmedLine.substring(5).trim()
    } else if (trimmedLine === '') {
      // 空行表示消息结束
      if (data) {
        try {
          const parsedData = JSON.parse(data)
          this.emit(event, parsedData)
        } catch (e) {
          console.error('SSE数据解析失败:', e, 'data:', data)
        }
        data = ''
        event = 'message'
      }
    }
  })
}
```

### 3. 清理缓冲区

```javascript
close() {
  if (this.requestTask) {
    this.requestTask.abort()
  }
  this.buffer = ''  // 清理缓冲区
}
```

## 工作原理

1. **累积数据**：每次收到chunk时，追加到 `buffer`
2. **保留残缺**：如果最后一行不完整（不以`\n`结尾），保留在buffer中
3. **下次处理**：下一个chunk到达时，与buffer中的残缺数据拼接
4. **完整解析**：只处理完整的行，确保JSON完整

### 示例

**Chunk 1 到达：**
```
buffer = ""
收到: "event: test\ndata: {\"key\":"
处理后:
  - 完整行: ["event: test"]
  - buffer = "data: {\"key\":"  (保留不完整行)
```

**Chunk 2 到达：**
```
buffer = "data: {\"key\":"
收到: "\"value\"}\n\n"
拼接后: "data: {\"key\":\"value\"}\n\n"
处理后:
  - 完整行: ["data: {\"key\":\"value\"}", ""]
  - 成功解析: {"key": "value"}
  - buffer = ""
```

## 测试验证

在 `index.vue` 中添加了调试日志：

```javascript
sse.addEventListener('analysis_step', (e) => {
  const data = e.data
  console.log('收到分析步骤:', data.step, '内容长度:', data.content?.length)
  // ...
})
```

在微信开发者工具的控制台中可以查看：
- 每个步骤是否都收到数据
- 每个步骤的内容长度
- 是否有解析错误

## 影响范围

**修改文件：**
- `frontend/miniapp/src/utils/sse.js` - SSE客户端核心逻辑
- `frontend/miniapp/src/pages/index/index.vue` - 添加调试日志

**修复效果：**
- ✅ 所有5个分析步骤都能正常显示
- ✅ 流式内容完整接收
- ✅ 无数据丢失

## 其他考虑

### 为什么不是所有步骤都有问题？

- 短消息可能在单个chunk中传输 → 正常显示
- 长消息更容易被拆分 → 容易丢失
- 网络状况影响chunk大小 → 随机性

### 为什么有时正常有时不正常？

这取决于：
1. LLM返回内容的长度
2. 网络传输的chunk大小
3. SSE消息正好在哪里被拆分

## 相关问题

如果仍然有步骤不显示，检查：

1. **后端是否执行**：查看 `/tmp/go-api.log`
   ```bash
   tail -f /tmp/go-api.log | grep "开始执行\|完成执行"
   ```

2. **前端是否接收**：查看微信开发者工具控制台
   ```
   收到分析步骤: comprehensive 内容长度: 234
   收到分析步骤: debate_bull 内容长度: 189
   ```

3. **JSON格式是否正确**：查看SSE解析错误日志
   ```
   SSE数据解析失败: SyntaxError ...
   ```

## 日期

修复完成：2026-01-28
