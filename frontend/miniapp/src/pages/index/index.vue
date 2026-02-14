<template>
  <view class="container">
    <!-- 搜索框 -->
    <view class="search-box">
      <input
        v-model="stockCode"
        class="input"
        placeholder="输入股票代码，如: 000630"
        :disabled="analyzing"
      />
      <button
        class="btn-analyze"
        @click="startAnalyze"
        :disabled="analyzing || !stockCode"
      >
        {{ analyzing ? '分析中...' : '开始分析' }}
      </button>
    </view>

    <!-- 进度条 -->
    <view v-if="analyzing" class="progress-bar">
      <view class="progress-inner" :style="{ width: progress + '%' }"></view>
      <text class="progress-text">{{ progress }}%</text>
    </view>

    <!-- 分析结果 -->
    <view v-if="results.length > 0" class="results">
      <view
        v-for="(result, index) in results"
        :key="index"
        class="result-card"
        :class="result.step"
        :id="'card-' + index"
      >
        <view class="card-header" @click="toggleExpand(index)">
          <text class="card-icon">{{ getIcon(result.step) }}</text>
          <text class="card-title">{{ result.role }}</text>
          <view class="header-right">
            <text v-if="result.streaming" class="typing-indicator">正在生成...</text>
            <text class="expand-icon" :class="{ expanded: result.expanded }">▼</text>
          </view>
        </view>
        <view v-show="result.expanded" class="card-content">
          <text class="content-text">{{ result.content }}</text>
          <view v-if="result.streaming" class="cursor-blink">▋</view>
        </view>
      </view>
    </view>

    <!-- 错误提示 -->
    <view v-if="error" class="error-box">
      <text>{{ error }}</text>
      <button @click="error = ''">关闭</button>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { SSEClient } from '../../utils/sse.js'
import { stockApi } from '../../api/stock.js'

const stockCode = ref('')
const analyzing = ref(false)
const progress = ref(0)
const results = ref([])
const error = ref('')
const currentStreamingStep = ref(null)  // 当前正在流式输出的步骤

// 开始分析
const startAnalyze = async () => {
  if (!stockCode.value) {
    uni.showToast({ title: '请输入股票代码', icon: 'none' })
    return
  }

  // 重置状态
  analyzing.value = true
  progress.value = 0
  results.value = []
  error.value = ''

  try {
    const api = stockApi.analyze(stockCode.value)
    const sse = new SSEClient(api.url, api.data)

    // 监听进度事件
    sse.addEventListener('progress', (e) => {
      const data = e.data  // 已经是对象，无需再次JSON.parse
      progress.value = data.progress
    })

    // 监听分析步骤
    sse.addEventListener('analysis_step', (e) => {
      const data = e.data  // 已经是对象，无需再次JSON.parse

      // 详细调试日志
      console.log('=====================')
      console.log('收到SSE事件 analysis_step')
      console.log('步骤:', data.step)
      console.log('角色:', data.role)
      console.log('内容片段:', data.content)
      console.log('内容长度:', data.content?.length)
      console.log('进度:', data.progress)
      console.log('当前results数量:', results.value.length)

      // 查找是否已有该步骤
      const existingIndex = results.value.findIndex(r => r.step === data.step)
      console.log('已存在索引:', existingIndex)

      if (existingIndex >= 0) {
        // 追加内容（流式）
        console.log('追加到已有步骤')
        results.value[existingIndex].content += data.content
        results.value[existingIndex].streaming = true
        console.log('追加后总长度:', results.value[existingIndex].content.length)
      } else {
        // 新增步骤（默认展开，标记为streaming）
        console.log('创建新步骤')
        const newResult = {
          step: data.step,
          role: data.role,
          content: data.content,
          expanded: true,  // 默认展开
          streaming: true  // 标记为正在流式输出
        }
        results.value.push(newResult)
        console.log('新步骤已添加，当前总数:', results.value.length)
        console.log('新步骤对象:', newResult)

        // 自动滚动到新步骤
        setTimeout(() => {
          const index = results.value.length - 1
          const query = uni.createSelectorQuery()
          query.select('#card-' + index).boundingClientRect()
          query.exec((res) => {
            if (res[0]) {
              uni.pageScrollTo({
                scrollTop: res[0].top,
                duration: 300
              })
            }
          })
        }, 100)
      }

      progress.value = data.progress
      currentStreamingStep.value = data.step

      // 输出当前所有结果的概况
      console.log('当前所有步骤:')
      results.value.forEach((r, i) => {
        console.log(`  [${i}] ${r.step} - ${r.role} - 长度:${r.content?.length || 0} - 展开:${r.expanded}`)
      })
      console.log('=====================')
    })

    // 监听步骤完成事件
    sse.addEventListener('step_completed', (e) => {
      const data = e.data
      console.log('收到step_completed事件, step:', data.step)
      const existingIndex = results.value.findIndex(r => r.step === data.step)
      if (existingIndex >= 0) {
        results.value[existingIndex].streaming = false
      }
      if (currentStreamingStep.value === data.step) {
        currentStreamingStep.value = null
      }
    })

    // 监听完成事件
    sse.addEventListener('done', () => {
      console.log('收到done事件')
      console.log('最终results:', results.value)

      // 移除所有streaming标记
      results.value.forEach(r => r.streaming = false)
      currentStreamingStep.value = null

      analyzing.value = false
      progress.value = 100
      uni.showToast({ title: '分析完成', icon: 'success' })
    })

    // 监听错误事件
    sse.addEventListener('error', (e) => {
      const data = e.data  // 已经是对象，无需再次JSON.parse
      error.value = data.error
      analyzing.value = false
    })

    // 开始连接
    await sse.connect()

  } catch (err) {
    error.value = '连接失败: ' + err.message
    analyzing.value = false
  }
}

// 切换展开/折叠
const toggleExpand = (index) => {
  results.value[index].expanded = !results.value[index].expanded
}

// 获取图标
const getIcon = (step) => {
  const icons = {
    'comprehensive': '📊',
    'debate_bull': '🐂',
    'debate_bear': '🐻',
    'trader': '💼',
    'final': '✅'
  }
  return icons[step] || '📝'
}
</script>

<style scoped>
.container {
  padding: 30rpx;
  min-height: 100vh;
}

.search-box {
  display: flex;
  gap: 20rpx;
  margin-bottom: 30rpx;
}

.input {
  flex: 1;
  height: 80rpx;
  padding: 0 30rpx;
  background: white;
  border-radius: 40rpx;
  border: 2rpx solid #e0e0e0;
}

.btn-analyze {
  width: 200rpx;
  height: 80rpx;
  line-height: 80rpx;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border-radius: 40rpx;
  border: none;
  font-size: 28rpx;
}

.btn-analyze:disabled {
  opacity: 0.6;
}

.progress-bar {
  position: relative;
  height: 40rpx;
  background: #f0f0f0;
  border-radius: 20rpx;
  overflow: hidden;
  margin-bottom: 30rpx;
}

.progress-inner {
  height: 100%;
  background: linear-gradient(90deg, #667eea 0%, #764ba2 100%);
  transition: width 0.3s;
}

.progress-text {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 24rpx;
  color: #333;
  font-weight: bold;
}

.results {
  display: flex;
  flex-direction: column;
  gap: 30rpx;
}

.result-card {
  background: white;
  border-radius: 20rpx;
  padding: 30rpx;
  box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.05);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 15rpx;
  margin-bottom: 20rpx;
  padding-bottom: 20rpx;
  border-bottom: 2rpx solid #f0f0f0;
  cursor: pointer;
  transition: background-color 0.2s;
}

.card-header:active {
  background-color: #f8f8f8;
  border-radius: 10rpx;
  margin: -10rpx;
  padding: 10rpx 10rpx 30rpx 10rpx;
}

.card-icon {
  font-size: 40rpx;
}

.card-title {
  flex: 1;
  font-size: 32rpx;
  font-weight: bold;
  color: #333;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 15rpx;
}

.typing-indicator {
  font-size: 22rpx;
  color: #1890ff;
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

.expand-icon {
  font-size: 24rpx;
  color: #999;
  transition: transform 0.3s ease;
  transform: rotate(0deg);
}

.expand-icon.expanded {
  transform: rotate(180deg);
}

.card-content {
  line-height: 1.8;
  overflow: hidden;
  transition: all 0.3s ease;
  position: relative;
}

.cursor-blink {
  display: inline-block;
  color: #1890ff;
  font-size: 32rpx;
  line-height: 1;
  animation: blink 1s step-end infinite;
  margin-left: 4rpx;
}

@keyframes blink {
  0%, 50% {
    opacity: 1;
  }
  51%, 100% {
    opacity: 0;
  }
}

.content-text {
  font-size: 28rpx;
  color: #666;
  white-space: pre-wrap;
}

.error-box {
  position: fixed;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 600rpx;
  padding: 40rpx;
  background: white;
  border-radius: 20rpx;
  box-shadow: 0 8rpx 24rpx rgba(0, 0, 0, 0.15);
  text-align: center;
}

.error-box text {
  display: block;
  margin-bottom: 30rpx;
  color: #f5222d;
}

.error-box button {
  width: 200rpx;
  height: 70rpx;
  line-height: 70rpx;
  background: #1890ff;
  color: white;
  border-radius: 35rpx;
  border: none;
}
</style>
