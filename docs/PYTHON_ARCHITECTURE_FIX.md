# Python 架构问题故障排除指南

## 问题描述

在 Apple Silicon (M系列) Mac 上启动 Python 服务时遇到架构不匹配错误：

```
ImportError: mach-o file, but is an incompatible architecture
(have 'x86_64', need 'arm64e' or 'arm64')
```

## 根本原因

- 系统是 ARM64 架构（Apple Silicon）
- 但部分 Python 包（numpy, pandas, cffi, curl-cffi）被安装成了 x86_64 版本
- 导致运行时无法加载动态库（.so 文件）

## 解决方案

### 1. 使用 ARM64 模式强制重装依赖

```bash
# 卸载 x86_64 版本
arch -arm64 python3 -m pip uninstall -y numpy pandas cffi curl-cffi

# 重新安装 ARM64 版本
arch -arm64 python3 -m pip install numpy pandas cffi

# 为 curl-cffi 设置 SSL 证书路径后安装
export SSL_CERT_FILE=$(python3 -c "import certifi; print(certifi.where())")
export CURL_CA_BUNDLE=$SSL_CERT_FILE
arch -arm64 python3 -m pip install curl-cffi --no-cache-dir
```

### 2. 验证安装

```bash
# 测试所有依赖是否正常导入
arch -arm64 python3 -c "import numpy; import pandas; import akshare; print('✅ 所有依赖导入成功')"
```

### 3. 启动服务

```bash
# 使用 ARM64 模式启动 Python 服务
arch -arm64 python3 backend/python-analysis/app.py
```

## 已完成的修复

✅ 重新安装所有依赖为 ARM64 版本：
- numpy: `macosx_14_0_arm64`
- pandas: `macosx_11_0_arm64`
- cffi: `macosx_11_0_arm64`
- curl-cffi: `macosx_14_0_arm64`

✅ 更新启动脚本 `scripts/start-dev.sh` 使用 `arch -arm64 python3`

## 检查架构版本

### 检查 Python 二进制文件

```bash
file $(which python3)
```

应该看到包含 `arm64` 的输出。

### 检查已安装包的架构

```bash
# 查看 numpy 的 .so 文件
file $(python3 -c "import numpy; print(numpy.__file__.replace('__init__.py', '_core/_multiarray_umath*.so'))")
```

正确的输出应包含 `arm64`。

## 为什么会出现这个问题？

1. **Rosetta 2 模拟**: 如果在 Rosetta 2 下运行终端，pip 会默认安装 x86_64 版本
2. **旧的 pip 缓存**: 之前在 x86_64 模式下安装的缓存
3. **混合环境**: 系统中同时存在 x86_64 和 ARM64 的 Python 环境

## 预防措施

1. **始终使用 `arch -arm64`**: 在 Apple Silicon Mac 上安装 Python 包时
   ```bash
   arch -arm64 python3 -m pip install <package>
   ```

2. **检查终端架构**: 确认当前终端是原生 ARM64
   ```bash
   uname -m  # 应该显示 arm64 (如果是原生) 或 x86_64 (如果在 Rosetta 下)
   ```

3. **使用虚拟环境**: 在 ARM64 模式下创建虚拟环境
   ```bash
   arch -arm64 python3 -m venv venv
   source venv/bin/activate
   ```

## 相关资源

- [NumPy 架构问题故障排除](https://numpy.org/devdocs/user/troubleshooting-importerror.html)
- [Apple Silicon Python 环境配置](https://developer.apple.com/documentation/apple-silicon)
- [Rosetta 2 文档](https://support.apple.com/en-us/HT211861)

## 快速诊断命令

```bash
# 检查当前架构
uname -m

# 检查 Python 架构
file $(which python3)

# 测试 numpy 导入
arch -arm64 python3 -c "import numpy; print('✅ NumPy OK')"

# 测试完整依赖链
arch -arm64 python3 -c "
import numpy
import pandas
import akshare
print('✅ 所有依赖正常')
"
```

## 日期

修复完成：2026-01-28
