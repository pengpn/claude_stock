#!/bin/bash

echo "=== 启动开发环境 ==="

# 检查.env文件
if [ ! -f .env ]; then
  echo "错误: 未找到.env文件，请从.env.example复制并配置"
  exit 1
fi

# 启动Python服务
echo "1. 启动Python分析服务..."
cd backend/python-analysis
python app.py &
PYTHON_PID=$!
echo "   Python服务 PID: $PYTHON_PID"
cd ../..

sleep 2

# 启动Go服务
echo "2. 启动Go API服务..."
cd backend/go-api
go run cmd/main.go &
GO_PID=$!
echo "   Go服务 PID: $GO_PID"
cd ../..

sleep 2

# 启动前端
echo "3. 启动小程序编译..."
cd frontend/miniapp
npm run dev:mp-weixin &
FRONTEND_PID=$!
echo "   前端编译 PID: $FRONTEND_PID"
cd ../..

# 保存PID
echo "$PYTHON_PID" > .dev.pid
echo "$GO_PID" >> .dev.pid
echo "$FRONTEND_PID" >> .dev.pid

echo ""
echo "=== 开发环境启动成功 ==="
echo "Python服务: http://localhost:5000"
echo "Go API: http://localhost:8080"
echo "请在微信开发者工具中打开: frontend/miniapp/dist/dev/mp-weixin"
echo ""
echo "停止服务: ./scripts/stop-dev.sh"
