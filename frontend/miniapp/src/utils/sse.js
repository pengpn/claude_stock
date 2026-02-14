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
    this.buffer = ''  // 添加缓冲区处理跨chunk数据
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
    // 将新数据追加到缓冲区
    this.buffer += text

    // 按行分割，保留最后一个不完整的行
    const lines = this.buffer.split('\n')

    // 如果缓冲区不是以\n结尾，最后一行是不完整的，需要保留
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
        // 空行表示一条消息结束
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
    this.buffer = ''  // 清理缓冲区
  }
}
