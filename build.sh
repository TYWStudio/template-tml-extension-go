#!/usr/bin/env bash
# 构建 Go 扩展并打包为 .tmlext
set -e
mkdir -p bin

if ! command -v go >/dev/null 2>&1; then
    echo "未找到 Go 工具链，请先安装：https://go.dev/dl/" >&2
    exit 1
fi

# SDK 来自 TML 仓库（首次构建会拉一次，之后走模块缓存，可离线）
go mod tidy

# CGO 关掉：产物是纯静态可执行，换机器也能跑
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/tml_ext ./src
chmod +x bin/tml_ext
python3 build.py
echo "完成：template-tml-extension-go.tmlext"
