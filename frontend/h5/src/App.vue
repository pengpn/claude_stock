<template>
  <div class="terminal" :class="{ light: !isDark }">
    <!-- Terminal Grid Background -->
    <div class="grid-overlay"></div>

    <div class="terminal-container">
      <!-- Terminal Header Bar -->
      <header class="terminal-header">
        <div class="header-left">
          <div class="terminal-dots">
            <span class="dot"></span>
            <span class="dot"></span>
            <span class="dot"></span>
          </div>
          <div class="terminal-title">
            <span class="title-bracket">[</span>
            <span class="title-text">股票分析终端</span>
            <span class="title-bracket">]</span>
          </div>
        </div>
        <div class="header-right">
          <div class="system-time">{{ currentTime }}</div>
          <button class="theme-switch" @click="toggleTheme" :title="isDark ? '浅色模式' : '深色模式'">
            <span v-if="isDark">☀</span>
            <span v-else>☾</span>
          </button>
        </div>
      </header>

      <!-- Command Input Section -->
      <section class="command-section">
        <div class="section-header">
          <span class="section-label">// 查询输入</span>
          <div class="section-line"></div>
        </div>

        <div class="command-wrapper">
          <div class="command-prompt">
            <span class="prompt-symbol">$</span>
            <span class="prompt-text">分析 --股票代码=</span>
          </div>
          <input
            v-model="stockCode"
            class="command-input"
            placeholder="600519"
            :disabled="analyzing"
            @keyup.enter="startAnalyze"
            @input="validateStockCode"
            maxlength="6"
          />
          <button
            class="execute-btn"
            @click="startAnalyze"
            :disabled="analyzing || !stockCode || !isValidCode"
          >
            <span v-if="!analyzing">[执行]</span>
            <span v-else class="loading-text">[运行中...]</span>
          </button>
        </div>

        <!-- History Bar -->
        <div v-if="history.length > 0 && !analyzing" class="history-bar">
          <span class="history-prefix">最近:</span>
          <button
            v-for="code in history"
            :key="code"
            class="history-item"
            @click="stockCode = code; startAnalyze()"
          >
            {{ code }}
          </button>
        </div>

        <!-- Progress Bar -->
        <div v-if="analyzing" class="progress-bar">
          <div class="progress-label">
            <span>处理中</span>
            <span class="progress-percent">{{ progress }}%</span>
          </div>
          <div class="progress-track">
            <div class="progress-fill" :style="{ width: progress + '%' }"></div>
          </div>
        </div>
      </section>

      <!-- Analysis Output -->
      <section v-if="results.length > 0" class="output-section">
        <div class="section-header">
          <span class="section-label">// 分析输出</span>
          <div class="section-line"></div>
        </div>

        <div class="output-grid">
          <div
            v-for="(result, index) in results"
            :key="index"
            class="output-block"
            :class="[`block-${result.step}`, { active: result.streaming }]"
          >
            <div class="block-header" @click="toggleExpand(index)">
              <div class="block-meta">
                <span class="block-index">[{{ String(index + 1).padStart(2, '0') }}]</span>
                <span class="block-type">{{ getStepLabel(result.step) }}</span>
                <span class="block-title">{{ result.role }}</span>
              </div>
              <div class="block-controls">
                <span v-if="result.streaming" class="status-indicator streaming">
                  <span class="blink">●</span> 流式传输中
                </span>
                <span v-else class="status-indicator complete">✓ 已完成</span>
                <button class="collapse-btn" :class="{ collapsed: !result.expanded }">
                  {{ result.expanded ? '[-]' : '[+]' }}
                </button>
              </div>
            </div>

            <transition name="slide">
              <div v-show="result.expanded" class="block-content">
                <div class="content-inner">
                  <div class="content-body" v-html="formatContent(result.content)"></div>
                  <span v-if="result.streaming" class="cursor-blink">█</span>
                </div>
              </div>
            </transition>
          </div>
        </div>
      </section>

      <!-- Empty State -->
      <section v-if="!analyzing && results.length === 0" class="empty-section">
        <div class="empty-content">
          <pre class="ascii-art">
    ╔═══════════════════════════════════╗
    ║      股票分析终端 v2.0           ║
    ║   ─────────────────────────────   ║
    ║      等待输入                     ║
    ╚═══════════════════════════════════╝
          </pre>
          <div class="empty-info">
            <p class="empty-desc">输入6位A股股票代码开始分析</p>
            <div class="example-codes">
              <span class="example-label">示例:</span>
              <button class="example-code" @click="stockCode = '600519'; startAnalyze()">600519</button>
              <button class="example-code" @click="stockCode = '000001'; startAnalyze()">000001</button>
              <button class="example-code" @click="stockCode = '600036'; startAnalyze()">600036</button>
            </div>
          </div>
        </div>
      </section>

      <!-- Error Modal -->
      <transition name="fade">
        <div v-if="error" class="error-overlay" @click="error = ''">
          <div class="error-box" @click.stop>
            <div class="error-header">
              <span class="error-icon">[!]</span>
              <span class="error-title">错误</span>
            </div>
            <div class="error-body">
              <pre class="error-message">{{ error }}</pre>
            </div>
            <button class="error-dismiss" @click="error = ''">[关闭]</button>
          </div>
        </div>
      </transition>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'

const stockCode = ref('')
const analyzing = ref(false)
const progress = ref(0)
const results = ref([])
const error = ref('')
const isDark = ref(true)
const isValidCode = ref(true)
const history = ref([])
const currentTime = ref('')
let timeInterval = null

// Update system time
const updateTime = () => {
  const now = new Date()
  currentTime.value = now.toLocaleTimeString('en-US', {
    hour12: false,
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

// Theme toggle
const toggleTheme = () => {
  isDark.value = !isDark.value
  localStorage.setItem('terminal-theme', isDark.value ? 'dark' : 'light')
}

// Load theme and history
onMounted(() => {
  const savedTheme = localStorage.getItem('terminal-theme')
  isDark.value = savedTheme !== 'light'

  const savedHistory = localStorage.getItem('stockHistory')
  if (savedHistory) {
    history.value = JSON.parse(savedHistory).slice(0, 5)
  }

  updateTime()
  timeInterval = setInterval(updateTime, 1000)
})

onUnmounted(() => {
  if (timeInterval) clearInterval(timeInterval)
})

// Validate stock code
const validateStockCode = () => {
  const code = stockCode.value.trim()
  isValidCode.value = /^[0-9]{6}$/.test(code) || code === ''
}

// Save history
const saveHistory = (code) => {
  const newHistory = [code, ...history.value.filter(c => c !== code)].slice(0, 5)
  history.value = newHistory
  localStorage.setItem('stockHistory', JSON.stringify(newHistory))
}

// Get step label
const getStepLabel = (step) => {
  const labels = {
    'comprehensive': '基本面分析',
    'debate_bull': '多头观点',
    'debate_bear': '空头观点',
    'trader': '交易信号',
    'final': '综合结论'
  }
  return labels[step] || '分析中'
}

// Start analysis
const startAnalyze = async () => {
  if (!stockCode.value || analyzing.value || !isValidCode.value) return

  analyzing.value = true
  progress.value = 0
  results.value = []
  error.value = ''
  saveHistory(stockCode.value)

  try {
    const response = await fetch('/api/v1/analyze', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ code: stockCode.value })
    })

    if (!response.ok) {
      throw new Error(`HTTP错误: ${response.status}`)
    }

    const reader = response.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      let currentEvent = ''
      for (const line of lines) {
        if (line.startsWith('event: ')) {
          currentEvent = line.substring(7).trim()
        } else if (line.startsWith('data: ')) {
          const dataStr = line.substring(6)
          try {
            const data = JSON.parse(dataStr)
            handleSSEData(currentEvent, data)
          } catch (e) {
            console.warn('[PARSE_ERROR]', e)
          }
          currentEvent = ''
        }
      }
    }

    results.value.forEach(r => r.streaming = false)
    analyzing.value = false
    progress.value = 100

  } catch (err) {
    error.value = '连接失败: ' + err.message
    analyzing.value = false
  }
}

// Handle SSE data
const handleSSEData = (eventType, data) => {
  if (eventType === 'progress') {
    if (data.progress !== undefined) {
      progress.value = data.progress
    }
    return
  }

  if (eventType === 'error') {
    error.value = data.error || '未知错误'
    analyzing.value = false
    return
  }

  if (eventType === 'step_completed') {
    const existingIndex = results.value.findIndex(r => r.step === data.step)
    if (existingIndex >= 0) {
      results.value[existingIndex].streaming = false
    }
    return
  }

  if (eventType === 'done') {
    results.value.forEach(r => r.streaming = false)
    analyzing.value = false
    progress.value = 100
    return
  }

  if (eventType === 'analysis_step') {
    const existingIndex = results.value.findIndex(r => r.step === data.step)

    if (existingIndex >= 0) {
      results.value[existingIndex].content += data.content
      results.value[existingIndex].streaming = true
    } else {
      results.value.push({
        step: data.step,
        role: data.role,
        content: data.content,
        expanded: true,
        streaming: true
      })
    }

    if (data.progress) {
      progress.value = data.progress
    }
  }
}

// Toggle expand/collapse
const toggleExpand = (index) => {
  results.value[index].expanded = !results.value[index].expanded
}

// Format content
const formatContent = (content) => {
  if (!content) return ''
  return content
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/\n/g, '<br>')
}
</script>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=IBM+Plex+Mono:wght@400;500;600;700&display=swap');

/* Terminal Color Variables */
.terminal {
  --term-bg: #0a0e27;
  --term-surface: #141b3a;
  --term-border: #1e2847;
  --term-text: #e2e8f0;
  --term-text-dim: #94a3b8;
  --term-text-muted: #64748b;
  --term-accent: #fbbf24;
  --term-success: #10b981;
  --term-danger: #ef4444;
  --term-warning: #f59e0b;
  --term-info: #3b82f6;
}

.terminal.light {
  --term-bg: #f5f5f0;
  --term-surface: #ffffff;
  --term-border: #d4d4d0;
  --term-text: #1e293b;
  --term-text-dim: #475569;
  --term-text-muted: #64748b;
  --term-accent: #d97706;
  --term-success: #059669;
  --term-danger: #dc2626;
  --term-warning: #ea580c;
  --term-info: #2563eb;
}

/* Base Layout */
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

.terminal {
  min-height: 100vh;
  background: var(--term-bg);
  color: var(--term-text);
  font-family: 'IBM Plex Mono', 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.6;
  position: relative;
  transition: background 0.3s, color 0.3s;
}

.grid-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-image:
    linear-gradient(var(--term-border) 1px, transparent 1px),
    linear-gradient(90deg, var(--term-border) 1px, transparent 1px);
  background-size: 20px 20px;
  opacity: 0.3;
  pointer-events: none;
  z-index: 0;
}

.terminal-container {
  max-width: 1400px;
  margin: 0 auto;
  padding: 20px;
  position: relative;
  z-index: 1;
}

/* Terminal Header */
.terminal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  background: var(--term-surface);
  border: 2px solid var(--term-border);
  margin-bottom: 24px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.terminal-dots {
  display: flex;
  gap: 8px;
}

.dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: var(--term-danger);
}

.dot:nth-child(2) {
  background: var(--term-warning);
}

.dot:nth-child(3) {
  background: var(--term-success);
}

.terminal-title {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 16px;
  font-weight: 600;
}

.title-bracket {
  color: var(--term-accent);
}

.title-text {
  color: var(--term-text);
  letter-spacing: 1px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.system-time {
  font-size: 14px;
  color: var(--term-success);
  font-weight: 500;
  letter-spacing: 1px;
}

.theme-switch {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: 2px solid var(--term-border);
  color: var(--term-accent);
  cursor: pointer;
  font-size: 18px;
  transition: all 0.2s;
}

.theme-switch:hover {
  background: var(--term-border);
  border-color: var(--term-accent);
}

/* Command Section */
.command-section {
  margin-bottom: 24px;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.section-label {
  color: var(--term-text-dim);
  font-size: 12px;
  letter-spacing: 1px;
  white-space: nowrap;
}

.section-line {
  flex: 1;
  height: 2px;
  background: var(--term-border);
}

.command-wrapper {
  display: flex;
  align-items: center;
  gap: 12px;
  background: var(--term-surface);
  border: 2px solid var(--term-border);
  padding: 12px 16px;
  margin-bottom: 12px;
}

.command-prompt {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--term-text-dim);
  white-space: nowrap;
}

.prompt-symbol {
  color: var(--term-accent);
  font-weight: 700;
}

.command-input {
  flex: 1;
  background: transparent;
  border: none;
  outline: none;
  color: var(--term-text);
  font-family: inherit;
  font-size: 14px;
  padding: 4px 8px;
}

.command-input::placeholder {
  color: var(--term-text-muted);
}

.command-input:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.execute-btn {
  padding: 8px 16px;
  background: var(--term-accent);
  color: var(--term-bg);
  border: none;
  font-family: inherit;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  letter-spacing: 0.5px;
}

.execute-btn:hover:not(:disabled) {
  background: var(--term-success);
}

.execute-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.loading-text {
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

/* History Bar */
.history-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  padding: 8px 12px;
  background: var(--term-surface);
  border: 2px solid var(--term-border);
  margin-bottom: 12px;
}

.history-prefix {
  color: var(--term-text-dim);
  font-size: 12px;
  letter-spacing: 1px;
}

.history-item {
  padding: 4px 12px;
  background: transparent;
  border: 1px solid var(--term-border);
  color: var(--term-text-dim);
  font-family: inherit;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.history-item:hover {
  background: var(--term-border);
  color: var(--term-accent);
  border-color: var(--term-accent);
}

/* Progress Bar */
.progress-bar {
  background: var(--term-surface);
  border: 2px solid var(--term-border);
  padding: 12px 16px;
}

.progress-label {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-size: 12px;
  color: var(--term-text-dim);
  letter-spacing: 1px;
}

.progress-percent {
  color: var(--term-accent);
  font-weight: 600;
}

.progress-track {
  height: 4px;
  background: var(--term-border);
  position: relative;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: var(--term-accent);
  transition: width 0.3s ease;
  position: relative;
}

.progress-fill::after {
  content: '';
  position: absolute;
  top: 0;
  right: 0;
  width: 20px;
  height: 100%;
  background: var(--term-success);
  animation: scan 1s linear infinite;
}

@keyframes scan {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(100%); }
}

/* Output Section */
.output-section {
  margin-bottom: 24px;
}

.output-grid {
  display: grid;
  gap: 16px;
}

.output-block {
  background: var(--term-surface);
  border: 2px solid var(--term-border);
  transition: border-color 0.3s;
}

.output-block.active {
  border-color: var(--term-accent);
}

.block-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  cursor: pointer;
  user-select: none;
  border-bottom: 2px solid var(--term-border);
}

.block-header:hover {
  background: rgba(251, 191, 36, 0.05);
}

.block-meta {
  display: flex;
  align-items: center;
  gap: 12px;
}

.block-index {
  color: var(--term-text-dim);
  font-size: 12px;
}

.block-type {
  color: var(--term-accent);
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 1px;
}

.block-title {
  color: var(--term-text);
  font-size: 14px;
  font-weight: 500;
}

.block-controls {
  display: flex;
  align-items: center;
  gap: 12px;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  letter-spacing: 1px;
  font-weight: 600;
}

.status-indicator.streaming {
  color: var(--term-info);
}

.status-indicator.complete {
  color: var(--term-success);
}

.blink {
  animation: blink 1s ease-in-out infinite;
}

@keyframes blink {
  0%, 50%, 100% { opacity: 1; }
  25%, 75% { opacity: 0.3; }
}

.collapse-btn {
  background: transparent;
  border: 1px solid var(--term-border);
  color: var(--term-text-dim);
  font-family: inherit;
  font-size: 13px;
  padding: 4px 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.collapse-btn:hover {
  background: var(--term-border);
  color: var(--term-accent);
}

.collapse-btn.collapsed {
  color: var(--term-text-muted);
}

.block-content {
  overflow: hidden;
}

.content-inner {
  padding: 16px;
}

.content-body {
  color: var(--term-text-dim);
  line-height: 1.8;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.content-body :deep(strong) {
  color: var(--term-text);
  font-weight: 600;
}

.cursor-blink {
  display: inline-block;
  color: var(--term-accent);
  animation: blink 1s step-end infinite;
  margin-left: 2px;
}

/* Slide Transition */
.slide-enter-active,
.slide-leave-active {
  transition: all 0.3s ease;
  max-height: 2000px;
}

.slide-enter-from,
.slide-leave-to {
  max-height: 0;
  opacity: 0;
}

/* Empty State */
.empty-section {
  padding: 60px 20px;
}

.empty-content {
  text-align: center;
}

.ascii-art {
  color: var(--term-accent);
  font-size: 12px;
  line-height: 1.4;
  margin-bottom: 24px;
  display: inline-block;
}

.empty-info {
  margin-top: 24px;
}

.empty-desc {
  color: var(--term-text-dim);
  font-size: 14px;
  margin-bottom: 20px;
}

.example-codes {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  flex-wrap: wrap;
}

.example-label {
  color: var(--term-text-muted);
  font-size: 12px;
  letter-spacing: 1px;
}

.example-code {
  padding: 6px 16px;
  background: transparent;
  border: 1px solid var(--term-border);
  color: var(--term-text-dim);
  font-family: inherit;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.example-code:hover {
  background: var(--term-accent);
  color: var(--term-bg);
  border-color: var(--term-accent);
}

/* Error Modal */
.error-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
}

.error-box {
  background: var(--term-surface);
  border: 2px solid var(--term-danger);
  padding: 24px;
  max-width: 500px;
  width: 100%;
}

.error-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 2px solid var(--term-border);
}

.error-icon {
  color: var(--term-danger);
  font-size: 20px;
  font-weight: 700;
}

.error-title {
  color: var(--term-danger);
  font-size: 16px;
  font-weight: 600;
  letter-spacing: 1px;
}

.error-body {
  margin-bottom: 20px;
}

.error-message {
  color: var(--term-text-dim);
  font-size: 13px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.error-dismiss {
  width: 100%;
  padding: 10px;
  background: var(--term-danger);
  color: white;
  border: none;
  font-family: inherit;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  letter-spacing: 0.5px;
}

.error-dismiss:hover {
  background: var(--term-accent);
  color: var(--term-bg);
}

/* Fade Transition */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* Responsive Design */
@media (max-width: 768px) {
  .terminal-container {
    padding: 12px;
  }

  .terminal-header {
    padding: 12px;
  }

  .terminal-title {
    font-size: 12px;
  }

  .title-text {
    display: none;
  }

  .system-time {
    font-size: 12px;
  }

  .command-wrapper {
    flex-direction: column;
    align-items: stretch;
  }

  .execute-btn {
    width: 100%;
  }

  .block-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }

  .block-controls {
    width: 100%;
    justify-content: space-between;
  }
}

@media (max-width: 480px) {
  .terminal-header {
    padding: 8px;
  }

  .header-left {
    gap: 8px;
  }

  .header-right {
    gap: 8px;
  }

  .dot {
    width: 8px;
    height: 8px;
  }

  .theme-switch {
    width: 32px;
    height: 32px;
    font-size: 14px;
  }

  .command-section {
    margin-bottom: 16px;
  }

  .content-inner {
    padding: 12px;
  }

  .ascii-art {
    font-size: 10px;
  }
}
</style>
