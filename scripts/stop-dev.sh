#!/bin/bash

echo "=== 停止开发环境 ==="

if [ -f .dev.pid ]; then
  while read pid; do
    if kill -0 $pid 2>/dev/null; then
      echo "停止进程 $pid"
      kill $pid
    fi
  done < .dev.pid
  rm .dev.pid
  echo "所有服务已停止"
else
  echo "未找到运行中的服务"
fi
