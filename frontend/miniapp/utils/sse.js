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
