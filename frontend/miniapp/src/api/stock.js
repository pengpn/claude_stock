const API_BASE = 'http://localhost:8000/api/v1'

export const stockApi = {
  // 分析股票（返回SSE连接的task）
  analyze(code) {
    return {
      url: `${API_BASE}/analyze`,
      data: { code }
    }
  }
}
