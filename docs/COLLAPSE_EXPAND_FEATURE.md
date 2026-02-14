# 分析步骤折叠/展开功能

## 功能说明

为每个分析步骤的卡片添加了折叠/展开功能，用户可以点击卡片头部来切换内容的显示/隐藏状态。

## 用户体验

### 默认状态
- 所有分析步骤**默认展开**，显示完整内容
- 适合首次查看时快速浏览所有分析结果

### 交互方式
1. **点击卡片头部**（包括图标、标题、箭头区域）
2. 卡片内容会**平滑展开/收起**
3. 右侧箭头图标会**旋转180度**指示状态

### 视觉反馈
- **点击时**：卡片头部背景变灰（:active效果）
- **展开状态**：箭头向下 ▼
- **收起状态**：箭头向上 ▲（旋转180度）
- **动画效果**：300ms的平滑过渡

## 实现细节

### 1. 数据结构

每个分析结果对象包含 `expanded` 属性：

```javascript
{
  step: 'comprehensive',      // 步骤标识
  role: '综合分析',            // 角色名称
  content: '分析内容...',      // 分析内容
  expanded: true              // 展开状态（默认true）
}
```

### 2. 模板结构

```vue
<view class="card-header" @click="toggleExpand(index)">
  <text class="card-icon">📊</text>
  <text class="card-title">综合分析</text>
  <text class="expand-icon" :class="{ expanded: result.expanded }">▼</text>
</view>
<view v-show="result.expanded" class="card-content">
  <text class="content-text">{{ result.content }}</text>
</view>
```

**关键点：**
- `@click="toggleExpand(index)"` - 点击切换状态
- `:class="{ expanded: result.expanded }"` - 动态类名
- `v-show="result.expanded"` - 控制内容显示

### 3. 切换方法

```javascript
const toggleExpand = (index) => {
  results.value[index].expanded = !results.value[index].expanded
}
```

简单的布尔值取反，Vue响应式系统自动更新UI。

### 4. CSS样式

#### 卡片头部（可点击）

```css
.card-header {
  display: flex;
  align-items: center;
  gap: 15rpx;
  cursor: pointer;
  transition: background-color 0.2s;
}

.card-header:active {
  background-color: #f8f8f8;
  border-radius: 10rpx;
  margin: -10rpx;
  padding: 10rpx 10rpx 30rpx 10rpx;
}
```

#### 标题区域（占据剩余空间）

```css
.card-title {
  flex: 1;  /* 占据剩余空间，推动箭头到右侧 */
  font-size: 32rpx;
  font-weight: bold;
  color: #333;
}
```

#### 展开箭头（旋转动画）

```css
.expand-icon {
  font-size: 24rpx;
  color: #999;
  transition: transform 0.3s ease;
  transform: rotate(0deg);
}

.expand-icon.expanded {
  transform: rotate(180deg);  /* 展开时旋转180度 */
}
```

#### 内容区域（平滑过渡）

```css
.card-content {
  line-height: 1.8;
  overflow: hidden;
  transition: all 0.3s ease;  /* 所有属性平滑过渡 */
}
```

## 使用场景

### 场景1：查看特定步骤
用户只对某个特定分析感兴趣（如"最终决策"），可以折叠其他步骤，聚焦关键信息。

### 场景2：对比多个步骤
展开"多头观点"和"空头观点"进行对比，折叠其他步骤。

### 场景3：长内容管理
当分析内容很长时，折叠已读步骤，减少滚动，提升浏览效率。

### 场景4：截图分享
折叠部分内容后截图，只分享关键信息。

## 技术优势

### 1. 性能优化
- 使用 `v-show` 而非 `v-if`
- DOM元素保留在文档中，只是隐藏
- 避免频繁的DOM创建/销毁
- 适合展开/收起频繁切换的场景

### 2. 响应式
- 利用Vue 3 Composition API
- 自动追踪 `expanded` 状态变化
- 无需手动操作DOM

### 3. 动画流畅
- CSS transition 硬件加速
- 300ms过渡时长（适中）
- transform旋转比修改其他属性更高效

### 4. 用户体验
- 默认展开：首次查看友好
- 点击区域大：整个头部可点击
- 视觉反馈清晰：箭头旋转+背景变化
- 动画自然：不突兀

## 可能的改进

### 1. 全部展开/收起按钮

在结果列表顶部添加：

```vue
<view class="control-bar">
  <button @click="expandAll">全部展开</button>
  <button @click="collapseAll">全部收起</button>
</view>
```

```javascript
const expandAll = () => {
  results.value.forEach(r => r.expanded = true)
}

const collapseAll = () => {
  results.value.forEach(r => r.expanded = false)
}
```

### 2. 记住用户偏好

使用 `uni.setStorageSync` 保存用户的展开/收起偏好：

```javascript
const toggleExpand = (index) => {
  results.value[index].expanded = !results.value[index].expanded

  // 保存到本地存储
  const preferences = results.value.map(r => ({
    step: r.step,
    expanded: r.expanded
  }))
  uni.setStorageSync('expand-preferences', preferences)
}
```

### 3. 高度动画

当前使用 `v-show`，高度变化无动画。如需高度动画：

```vue
<transition name="expand">
  <view v-if="result.expanded" class="card-content">
    ...
  </view>
</transition>
```

```css
.expand-enter-active, .expand-leave-active {
  transition: all 0.3s ease;
  max-height: 1000rpx;
}

.expand-enter-from, .expand-leave-to {
  max-height: 0;
  opacity: 0;
}
```

但需要注意：
- `v-if` 会重新创建/销毁DOM（性能开销）
- 需要设置 `max-height`（难以精确预估）

### 4. 手势滑动

支持左右滑动手势折叠/展开：

```vue
<view
  @touchstart="handleTouchStart"
  @touchmove="handleTouchMove"
  @touchend="handleTouchEnd"
>
  ...
</view>
```

## 修改的文件

- `frontend/miniapp/src/pages/index/index.vue`
  - 模板：添加展开箭头，点击事件
  - 脚本：添加 `expanded` 属性，`toggleExpand` 方法
  - 样式：添加可点击样式、箭头旋转动画

## 兼容性

- ✅ 微信小程序
- ✅ Vue 3 Composition API
- ✅ CSS Flexbox
- ✅ CSS Transitions
- ✅ 所有现代浏览器

## 测试建议

1. **基础功能**
   - 点击每个卡片头部
   - 验证内容正确展开/收起
   - 验证箭头正确旋转

2. **边界情况**
   - 快速连续点击同一卡片
   - 分析进行中时点击（流式内容追加）
   - 内容为空时的显示

3. **性能**
   - 5个步骤全部展开/收起
   - 长内容（>1000字）的展开/收起
   - 低端设备上的流畅度

4. **视觉**
   - 点击区域是否足够大
   - 动画是否流畅自然
   - 不同屏幕尺寸下的显示

## 日期

功能添加：2026-01-28
