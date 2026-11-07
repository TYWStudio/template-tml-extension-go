// TML Extension API v1 —— Go 语言宿主 SDK（运行时：native / go）
//
// TYW Launcher 的扩展永远运行在启动器进程之外，仅通过 stdin/stdout 的
// NDJSON（每行一个 JSON）交换白名单方法。本文件是最小实现，只依赖 Go
// 标准库：
//
//   - 自动完成 initialize 握手；
//   - 封装 log / toast / host.info / instances.list / event.subscribe /
//     storage.get / storage.set / settings.get / settings.set；
//   - 接收宿主广播的事件，收到 shutdown 优雅退出。
//
// 可调用的能力由 manifest.json 的 permissions 决定，未申请的永远不会返回数据。
// 详见启动器「设置 → 其他 → 拓展」或拓展协议文档 extensions/README.md。
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// host 维护与启动器的 NDJSON 会话。
type host struct {
	in     *bufio.Scanner
	nextID int64
	stop   bool
}

func newHost() *host {
	sc := bufio.NewScanner(os.Stdin)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	return &host{in: sc, nextID: 2}
}

// send 向宿主写一行 JSON。
func (h *host) send(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(os.Stdout, "%s\n", b)
	return err
}

func toInt(v any) int64 {
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int64:
		return n
	case json.Number:
		i, _ := n.Int64()
		return i
	}
	return -1
}

// call 同步调用宿主能力，阻塞到拿到同 id 的响应。
func (h *host) call(method string, params map[string]any) (map[string]any, error) {
	id := h.nextID
	h.nextID++
	msg := map[string]any{"id": id, "method": method}
	if params != nil {
		msg["params"] = params
	}
	if err := h.send(msg); err != nil {
		return nil, err
	}
	for h.in.Scan() {
		line := strings.TrimSpace(h.in.Text())
		if line == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			continue
		}
		if rid, ok := m["id"]; ok && toInt(rid) == id {
			if e, isErr := m["error"]; isErr {
				return nil, fmt.Errorf("宿主返回错误：%v", e)
			}
			if r, ok := m["result"].(map[string]any); ok {
				return r, nil
			}
			return map[string]any{}, nil
		}
		h.handle(m)
		if h.stop {
			break
		}
	}
	return nil, io.EOF
}

// handle 处理一条来自宿主的消息；返回 false 表示应当退出。
func (h *host) handle(m map[string]any) bool {
	if _, isResp := m["result"]; isResp {
		return !h.stop
	}
	if _, isErr := m["error"]; isErr {
		return !h.stop
	}
	method, _ := m["method"].(string)
	id := toInt(m["id"])
	switch method {
	case "initialize":
		_ = h.send(map[string]any{"id": id, "result": map[string]any{"ready": true}})
	case "shutdown":
		h.stop = true
	case "event":
		if p, ok := m["params"].(map[string]any); ok {
			if t, ok := p["type"].(string); ok {
				fmt.Fprintf(os.Stderr, "[event] %s\n", t)
			}
		}
	default:
		// 宿主下发的自定义命令：模板一律确认
		_ = h.send(map[string]any{"id": id, "result": map[string]any{}})
	}
	return !h.stop
}

// ------------------------- 白名单能力封装 -------------------------

func (h *host) log(level, message string) {
	_, _ = h.call("log", map[string]any{"level": level, "message": message})
}

func (h *host) toast(message string) {
	_, _ = h.call("toast", map[string]any{"message": message})
}

func (h *host) hostInfo() map[string]any {
	r, _ := h.call("host.info", nil)
	return r
}

func (h *host) listInstances() map[string]any {
	r, _ := h.call("instances.list", nil)
	return r
}

func (h *host) subscribe(types []string) map[string]any {
	r, _ := h.call("event.subscribe", map[string]any{"types": types})
	return r
}

func (h *host) storageSet(key, value string) {
	_, _ = h.call("storage.set", map[string]any{"key": key, "value": value})
}

func (h *host) storageGet(key string) map[string]any {
	r, _ := h.call("storage.get", map[string]any{"key": key})
	return r
}

func (h *host) settingsSet(key, value string) {
	_, _ = h.call("settings.set", map[string]any{"key": key, "value": value})
}

func (h *host) settingsGet(key string) map[string]any {
	r, _ := h.call("settings.get", map[string]any{"key": key})
	return r
}

func main() {
	// 宿主下发的协议配置（环境变量）
	extID := os.Getenv("TML_EXT_ID")
	api := os.Getenv("TML_EXT_API")
	fmt.Fprintf(os.Stderr, "[Go 扩展] id=%s api=%s\n", extID, api)

	h := newHost()

	// 等待 initialize 握手
	ready := false
	for h.in.Scan() {
		line := strings.TrimSpace(h.in.Text())
		if line == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			continue
		}
		method, _ := m["method"].(string)
		h.handle(m)
		if method == "initialize" {
			ready = true
			break
		}
		if h.stop {
			break
		}
	}
	if !ready {
		// 没握手就断开（宿主没启动我们，或 stdin 被关掉）：直接退出，不执行业务逻辑
		fmt.Fprintln(os.Stderr, "[Go 扩展] 未收到 initialize 握手，退出")
		return
	}

	// ===================== 下面是扩展的业务逻辑示例 =====================
	h.log("info", "Go 扩展已启动（TML Extension API v1）")

	if info := h.hostInfo(); info != nil {
		if b, err := json.Marshal(info); err == nil {
			fmt.Fprintf(os.Stderr, "[host.info] %s\n", b)
		}
	}
	if list := h.listInstances(); list != nil {
		if b, err := json.Marshal(list); err == nil {
			fmt.Fprintf(os.Stderr, "[instances] %s\n", b)
		}
	}
	h.subscribe([]string{"launcher.started", "game.launched", "settings.changed"})

	h.storageSet("demo", "hello")
	fmt.Fprintf(os.Stderr, "[storage.get demo] %v\n", h.storageGet("demo"))
	h.settingsSet("foo", "bar")
	fmt.Fprintf(os.Stderr, "[settings.get foo] %v\n", h.settingsGet("foo"))

	h.toast("Go 扩展已就绪")

	// 保持运行，处理事件与命令直到 shutdown
	for h.in.Scan() {
		line := strings.TrimSpace(h.in.Text())
		if line == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			continue
		}
		if !h.handle(m) {
			break
		}
	}
	fmt.Fprintln(os.Stderr, "[Go 扩展] 已退出")
}
