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
      <!-- 综合分析 -->
      <view
        v-for="(result, index) in results"
        :key="index"
        class="result-card"
        :class="result.step"
      >
        <view class="card-header">
          <text class="card-icon">{{ getIcon(result.step) }}</text>
          <text class="card-title">{{ result.role }}</text>
        </view>
        <view class="card-content">
          <text class="content-text">{{ result.content }}</text>
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
      const data = JSON.parse(e.data)
      progress.value = data.progress
    })

    // 监听分析步骤
    sse.addEventListener('analysis_step', (e) => {
      const data = JSON.parse(e.data)

      // 查找是否已有该步骤
      const existingIndex = results.value.findIndex(r => r.step === data.step)

      if (existingIndex >= 0) {
        // 追加内容（流式）
        results.value[existingIndex].content += data.content
      } else {
        // 新增步骤
        results.value.push({
          step: data.step,
          role: data.role,
          content: data.content
        })
      }

      progress.value = data.progress
    })

    // 监听完成事件
    sse.addEventListener('done', () => {
      analyzing.value = false
      progress.value = 100
      uni.showToast({ title: '分析完成', icon: 'success' })
    })

    // 监听错误事件
    sse.addEventListener('error', (e) => {
      const data = JSON.parse(e.data)
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
}

.card-icon {
  font-size: 40rpx;
}

.card-title {
  font-size: 32rpx;
  font-weight: bold;
  color: #333;
}

.card-content {
  line-height: 1.8;
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
