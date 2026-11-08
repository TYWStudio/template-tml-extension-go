<p align="right">
  <a href="README.md">English</a> · <b>简体中文</b>
</p>

> 本仓库属于 **TML** 扩展生态；启动器仓库：<https://github.com/TYWStudio/TML>

# TML 扩展模板 · Go（template-tml-extension-go）

适用于 **TYW Launcher（TML）扩展 API v1** 的 **Go 语言**扩展模板。宿主（启动器）在独立进程里
拉起你编译好的静态可执行文件，双方仅通过 **stdin / stdout 的 NDJSON（每行一个 JSON）** 交换白名单方法。

SDK（会话、握手、按 id 分发、白名单能力）只在 **TML 仓库** 的
[`extensions/sdk/go`](https://github.com/TYWStudio/TML/tree/main/extensions/sdk/go) 维护，
本仓库通过 `go.mod` 引用它（`require .../extensions/sdk/go v1.0.0`），`src/main.go` 里只写业务逻辑。
**不要在模板里复制一份 SDK**：协议改动只会发生在 TML 仓库，模板升级依赖版本即可。

> 许可：**CC BY-NC 4.0**（禁止商业使用，署名）。见 `LICENSE`。

---

## 它是怎么跑起来的

1. 你在启动器「设置 → 其他 → 拓展」里导入打包好的 `*.tmlext`（本质是一个 zip）。
2. 启动器解压并启动包里的静态可执行文件（`manifest.json` 的 `entry`）。
3. 启动器先发一条 `initialize` 握手，扩展回 `{"ready":true}`。
4. 之后扩展通过 `id≥2` 的请求调用宿主能力（`log` / `host.info` / `storage` / `settings` …），
   宿主通过 `event` 把事件广播给扩展，并通过 `shutdown` 通知扩展优雅退出。

扩展**默认全部禁用**、需用户手动启用，且只能调用 `manifest.json` 的 `permissions` 中申请过的
能力——未申请的调用宿主不会返回数据。

## 协议（极简）

```
宿主 → 扩展: {"id":1,"method":"initialize","params":{...}}    扩展 → 宿主: {"id":1,"result":{"ready":true}}
扩展 → 宿主: {"id":2,"method":"log","params":{"level":"info","message":"hi"}}  宿主 → 扩展: {"id":2,"result":{}}
宿主 → 扩展: {"method":"event","params":{"type":"game.launched","data":{...}}}
宿主 → 扩展: {"method":"shutdown"}                                （扩展直接退出）
```

会话实现不在本仓库（见上面的 SDK），`src/main.go` 只处理 `Ready` / `OnEvent`。宿主还会通过环境变量下发配置：

| 变量 | 含义 |
| --- | --- |
| `TML_EXT_ID` | 扩展唯一 id |
| `TML_EXT_API` | API 版本（当前为 1） |
| `TML_EXT_PLATFORM` | 平台（windows / linux / macos） |
| `TML_EXT_LAUNCHER_VERSION` | 启动器版本 |
| `TML_EXT_CHANNEL` | 通道（Release / Snapshot / Beta） |
| `TML_EXT_LANGUAGE` | 界面语言 |
| `TML_EXT_PRIVATE` | 扩展私有数据目录 |
| `TML_EXT_NONCE` | 一次性随机数 |
| `TML_EXT_PERMISSIONS` | 已授予的能力（逗号分隔） |

---

## 构建

需要 **Go 1.21+** 与 Python 3（打包用）。SDK 是普通 Go 模块依赖，`build.sh` / `build.bat` 里会先跑
`go mod tidy`：首次构建从模块代理拉一次（国内可设 `GOPROXY=https://goproxy.cn,direct`），之后走本地缓存。

### Linux / macOS

```bash
./build.sh
# 产物：template-tml-extension-go.tmlext
```

### Windows

```bat
build.bat
 rem 产物：template-tml-extension-go.tmlext
```

两个脚本都会执行 `CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/tml_ext ./src`，再由
`build.py` 把 `manifest.json` + 源码 + 二进制打成 `.tmlext`。`CGO_ENABLED=0` 保证产物是纯静态
二进制，同一平台的任何机器都能直接跑。

### 交叉编译

Go 原生支持交叉编译，想给哪个平台出包就编哪个：

```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/tml_ext.exe ./src
GOOS=linux   GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/tml_ext ./src
GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/tml_ext ./src
```

## 测试

`tests/smoke.py` 扮演宿主：拉起编译好的二进制、完成握手、应答白名单调用、广播事件并发送
`shutdown`。CI 每次推送都会跑一遍。

```bash
python3 tests/check_manifest.py go     # 清单体检
python3 tests/smoke.py ./bin/tml_ext   # 协议冒烟测试
```

## 装进启动器

把生成的 `template-tml-extension-go.tmlext` 拿到启动器「设置 → 其他 → 拓展」点 **导入**，
再查看权限列表并启用，即可在启动器控制台看到本模板的调试输出。

## manifest.json 关键字段

| 字段 | 说明 |
| --- | --- |
| `runtime` | 填 `go`，宿主按原生可执行处理（与 `c` / `native` 同一条路径） |
| `format` | 必须为 `tmlext`（与后缀一致） |
| `entry` | 包内二进制路径，如 `bin/tml_ext`（Windows 下自动补 `.exe`） |
| `platforms` | 必须包含当前平台（`windows` / `linux` / `macos`） |
| `permissions` | 申请的能力白名单 |
| `subscribe` | 想接收的事件类型 |

## 目录结构

```
template-tml-extension-go/
├── manifest.json         # 扩展清单
├── go.mod                # 引用 TML 仓库的 SDK（go.sum 已锁定版本）
├── src/main.go           # 业务逻辑（Ready / OnEvent），SDK 不在本仓库
├── build.py              # 跨平台打包为 .tmlext
├── build.sh / build.bat  # 编译 + 打包
├── tests/                # 清单校验 + 宿主侧协议冒烟测试
├── README.md / README-CN.md
└── LICENSE               # CC BY-NC 4.0
```
