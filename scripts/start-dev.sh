#!/bin/bash

echo "=== 启动开发环境 ==="

# 获取项目根目录（脚本所在目录的父目录）
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"

# 检查.env文件
if [ ! -f "$PROJECT_ROOT/.env" ]; then
  echo "错误: 未找到.env文件，请从.env.example复制并配置"
  exit 1
fi

# 先停止已有服务，避免端口冲突
echo "0. 清理旧进程..."
"$SCRIPT_DIR/stop-dev.sh" 2>/dev/null
sleep 1

# 启动Python服务
echo "1. 启动Python分析服务..."
cd "$PROJECT_ROOT/backend/python-analysis" && arch -arm64 python3 app.py > /tmp/python-service.log 2>&1 &
PYTHON_PID=$!
echo "   Python服务 PID: $PYTHON_PID"

sleep 2

# 启动Go服务
echo "2. 启动Go API服务..."
cd "$PROJECT_ROOT/backend/go-api" && go run cmd/main.go > /tmp/go-api.log 2>&1 &
GO_PID=$!
echo "   Go服务 PID: $GO_PID"

sleep 2

# 启动前端
echo "3. 启动小程序编译..."
cd "$PROJECT_ROOT" && npm --prefix "$PROJECT_ROOT/frontend/miniapp" run dev:mp-weixin > /tmp/frontend.log 2>&1 &
FRONTEND_PID=$!
echo "   前端编译 PID: $FRONTEND_PID"

# 保存PID
echo "$PYTHON_PID" > "$PROJECT_ROOT/.dev.pid"
echo "$GO_PID" >> "$PROJECT_ROOT/.dev.pid"
echo "$FRONTEND_PID" >> "$PROJECT_ROOT/.dev.pid"

echo ""
echo "=== 开发环境启动成功 ==="
echo "Python服务: http://localhost:8001"
echo "Go API: http://localhost:8000"
echo "请在微信开发者工具中打开: $PROJECT_ROOT/frontend/miniapp/dist/dev/mp-weixin"
echo ""
echo "查看日志:"
echo "  Python: tail -f /tmp/python-service.log"
echo "  Go API: tail -f /tmp/go-api.log"
echo "  前端: tail -f /tmp/frontend.log"
echo ""
echo "停止服务: $SCRIPT_DIR/stop-dev.sh"
