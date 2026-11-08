@echo off
setlocal
if not exist bin mkdir bin
where go >nul 2>nul
if errorlevel 1 (
    echo 未找到 Go 工具链，请先安装：https://go.dev/dl/
    exit /b 1
)
rem SDK 来自 TML 仓库（首次构建会拉一次，之后走模块缓存，可离线）
go mod tidy
if errorlevel 1 exit /b 1
set CGO_ENABLED=0
go build -trimpath -ldflags="-s -w" -o bin\tml_ext.exe .\src
if errorlevel 1 exit /b 1
python build.py
echo 完成：template-tml-extension-go.tmlext
