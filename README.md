<p align="right">
  <b>English</b> · <a href="README-CN.md">简体中文</a>
</p>

# TML Extension Template · Go

A **Go** extension template for the **TYW Launcher (TML)** extension API v1. The host (launcher)
spawns your compiled static binary in a separate process and the two sides exchange allow-listed
methods over **NDJSON on stdin/stdout** (one JSON object per line).

The SDK (session, handshake, request dispatch, allow-listed capabilities) lives in the **TML
repository** at [`extensions/sdk/go`](https://github.com/TYWStudio/TML/tree/main/extensions/sdk/go).
This template references it from `go.mod` (`require .../extensions/sdk/go v1.0.0`) and keeps
`src/main.go` for your own logic — **never copy the SDK in here**.

> License: **CC BY-NC 4.0** (non-commercial, attribution). See `LICENSE`.
>
> Part of the **TML** ecosystem · launcher repository: <https://github.com/TYWStudio/TML>

---

## How it runs

1. You import the packaged `*.tmlext` (essentially a zip) via **Settings → Other → Extensions** in the launcher.
2. The launcher unpacks it and starts the static binary inside the package (`entry` in `manifest.json`).
3. The launcher sends an `initialize` handshake first; the extension replies `{"ready":true}`.
4. After that the extension calls host capabilities with requests carrying `id≥2` (`log` / `host.info` /
   `storage` / `settings` …); the host broadcasts events as `event` lines and asks the extension to
   exit gracefully with `shutdown`.

Extensions are **disabled by default**, must be enabled manually by the user, and may only call the
capabilities declared in the manifest's `permissions` — calls that were not requested return no data.

## Protocol (minimal)

```
Host → Ext: {"id":1,"method":"initialize","params":{...}}    Ext → Host: {"id":1,"result":{"ready":true}}
Ext → Host: {"id":2,"method":"log","params":{"level":"info","message":"hi"}}  Host → Ext: {"id":2,"result":{}}
Host → Ext: {"method":"event","params":{"type":"game.launched","data":{...}}}
Host → Ext: {"method":"shutdown"}                                (the extension just exits)
```

The session is implemented by the SDK, not by this repository; `src/main.go` only handles `Ready` /
`OnEvent`. The host also passes configuration
through environment variables:

| Variable | Meaning |
| --- | --- |
| `TML_EXT_ID` | Unique extension id |
| `TML_EXT_API` | API version (currently 1) |
| `TML_EXT_PLATFORM` | Platform (windows / linux / macos) |
| `TML_EXT_LAUNCHER_VERSION` | Launcher version |
| `TML_EXT_CHANNEL` | Channel (Release / Snapshot / Beta) |
| `TML_EXT_LANGUAGE` | UI language |
| `TML_EXT_PRIVATE` | Extension-private data directory |
| `TML_EXT_NONCE` | One-time nonce |
| `TML_EXT_PERMISSIONS` | Granted capabilities (comma-separated) |

---

## Build

Requires **Go 1.21+** and Python 3 (for packaging). The SDK is a normal module dependency:
`build.sh` / `build.bat` run `go mod tidy` first, which downloads it once (behind the Great Firewall
set `GOPROXY=https://goproxy.cn,direct`), then it comes from the local cache.

### Linux / macOS

```bash
./build.sh
# output: template-tml-extension-go.tmlext
```

### Windows

```bat
build.bat
 rem output: template-tml-extension-go.tmlext
```

Both scripts run `CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/tml_ext ./src`, then
`build.py` packs `manifest.json` + sources + the binary into the `.tmlext`. `CGO_ENABLED=0` keeps the
result a pure static binary, so the same package runs on any machine of that OS/arch.

### Cross-compiling

Go cross-compiles out of the box — build the binary for each platform you want to ship:

```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/tml_ext.exe ./src
GOOS=linux   GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/tml_ext ./src
GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/tml_ext ./src
```

## Test

`tests/smoke.py` plays the host role: it spawns the built binary, performs the handshake, answers the
allow-listed calls, broadcasts an event and sends `shutdown`. CI runs it on every push.

```bash
python3 tests/check_manifest.py go     # manifest sanity
python3 tests/smoke.py ./bin/tml_ext   # protocol smoke test
```

## Install into the launcher

Copy the generated `template-tml-extension-go.tmlext` into **Settings → Other → Extensions** and click
**Import**. Then review the permission list, enable it, and watch the extension's debug output in the
launcher console.

## manifest.json key fields

| Field | Description |
| --- | --- |
| `runtime` | `go` — the host treats it as a native executable (same path as `c` / `native`) |
| `format` | Must be `tmlext` (matches the file extension) |
| `entry` | Package-relative path to the binary, e.g. `bin/tml_ext` (`.exe` is appended automatically on Windows) |
| `platforms` | The current platform must be included (`windows` / `linux` / `macos`) |
| `permissions` | Requested capability allow-list |
| `subscribe` | Event types to receive |

## Layout

```
template-tml-extension-go/
├── manifest.json         # Extension manifest
├── go.mod                # std-only module
├── src/main.go           # your logic (Ready / OnEvent); the SDK lives in TML
├── build.py              # Cross-platform packaging into .tmlext
├── build.sh / build.bat  # Compile + package
├── tests/                # Manifest check + host-side protocol smoke test
├── README.md
└── LICENSE               # CC BY-NC 4.0
```
