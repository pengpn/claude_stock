#!/bin/bash

echo "=== 停止开发环境 ==="

stopped=0

# 方法1: 按端口杀进程（最可靠）
kill_by_port() {
  local port=$1
  local name=$2
  local pids=$(lsof -ti :$port 2>/dev/null)
  if [ -n "$pids" ]; then
    echo "停止 $name (端口 $port, PID: $pids)"
    echo "$pids" | xargs kill -9 2>/dev/null
    stopped=1
  fi
}

# 方法2: 按进程名杀进程（兜底）
kill_by_name() {
  local pattern=$1
  local name=$2
  local pids=$(pgrep -f "$pattern" 2>/dev/null)
  if [ -n "$pids" ]; then
    echo "停止 $name (PID: $pids)"
    echo "$pids" | xargs kill -9 2>/dev/null
    stopped=1
  fi
}

# 停止 Python 分析服务 (端口 8001)
kill_by_port 8001 "Python分析服务"

# 停止 Go API 服务 (端口 8000)
kill_by_port 8000 "Go API服务"

# 停止 H5 前端服务 (端口 3000)
kill_by_port 3000 "H5前端服务"

# 兜底：通过保存的 PID 文件清理残留进程树
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"
PID_FILE="$PROJECT_ROOT/.dev.pid"

if [ -f "$PID_FILE" ]; then
  while read pid; do
    if [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null; then
      # 杀掉该进程及其所有子进程
      pkill -P "$pid" 2>/dev/null
      kill -9 "$pid" 2>/dev/null
      echo "清理残留进程: $pid"
      stopped=1
    fi
  done < "$PID_FILE"
  rm -f "$PID_FILE"
fi

if [ "$stopped" -eq 1 ]; then
  echo "所有服务已停止"
else
  echo "未找到运行中的服务"
fi
