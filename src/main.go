// TML 扩展模板 · Go
//
// 业务代码只写在这个文件里。NDJSON 会话、initialize 握手、按 id 分发请求、
// 白名单能力（log / toast / host.info / instances.list / storage / settings）
// 全部来自 TML 仓库的 extensions/sdk/go —— 通过 go.mod 引用，不要复制一份进来。
//
// 宿主（TYW Launcher）把 .tmlext 解开后，直接把包里 bin/tml_ext 当独立进程拉起，
// 双方只用 stdin/stdout 通信；能调用什么由 manifest.json 的 permissions 决定。
package main

import (
	"encoding/json"

	tmlext "github.com/TYWStudio/TML/extensions/sdk/go"
)

// extension 是业务入口，两个方法都可以留空。
type extension struct{}

// Ready 在握手成功后调用一次，适合做初始化。
func (extension) Ready() {
	tmlext.Log("info", "扩展已启动（TML Extension API v1）")

	// 宿主信息：平台、架构、启动器版本、通道、语言
	if info := tmlext.HostInfo(); info != nil {
		if b, err := json.Marshal(info); err == nil {
			tmlext.Log("info", "host.info = "+string(b))
		}
	}

	// 只读实例列表（需要 manifest 里申请 instances.list）
	for _, inst := range tmlext.Instances() {
		if b, err := json.Marshal(inst); err == nil {
			tmlext.Log("info", "instance = "+string(b))
		}
	}

	// 订阅事件（manifest 的 subscribe 里也要写）
	_ = tmlext.Subscribe("launcher.started", "game.launched", "settings.changed")

	// 私有存储：跟着扩展自己的目录走，卸载就没了
	_ = tmlext.StorageSet("lastRun", tmlext.Env("launcher_version"))
	if v, ok := tmlext.StorageGet("lastRun"); ok {
		tmlext.Log("info", "上次运行版本："+v)
	}

	// 扩展设置：显示在启动器的扩展详情里
	_ = tmlext.SettingsSet("greeted", "1")
	if v, ok := tmlext.SettingsGet("greeted"); ok {
		tmlext.Log("info", "greeted = "+v)
	}

	tmlext.Toast("扩展已就绪")
}

// OnEvent 接收订阅到的事件（数据已由宿主脱敏）。
func (extension) OnEvent(e tmlext.Event) {
	switch e.Type {
	case "launcher.started":
		tmlext.Log("info", "启动器已就绪")
	case "game.launched":
		name, _ := e.Data["name"].(string)
		tmlext.Log("info", "开始游戏："+name)
	default:
		tmlext.Log("info", "事件 "+e.Type)
	}
}

func main() {
	tmlext.Run(extension{})
}
