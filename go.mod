module tmlext

go 1.21

// SDK 只在 TML 仓库维护，这里只引用、不复制。
// 升级：改版本号后 `go mod tidy`；SDK 源码在
// https://github.com/TYWStudio/TML/tree/main/extensions/sdk/go
require github.com/TYWStudio/TML/extensions/sdk/go v1.0.0
