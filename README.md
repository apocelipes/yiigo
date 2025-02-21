# Web开发脚手架

[![golang](https://img.shields.io/badge/Language-Go-green.svg?style=flat)](https://golang.org)
[![pkg.go.dev](https://img.shields.io/badge/dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/yiigo/yiigo)
[![Apache 2.0 license](http://img.shields.io/badge/license-Apache%202.0-brightgreen.svg)](http://opensource.org/licenses/apache2.0)

## yiigo

自动生成项目，支持 `HTTP` 和 `gRPC`，并同时支持创建「单应用」和「多应用」

```shell
go install github.com/yiigo/yiigo/cmd/yiigo@latest
```

进一步了解 👉 [详情](cmd/yiigo/README.md)

## protoc-gen-yiigo

使用 `proto` 定义API，基于 [chi](https://github.com/go-chi/chi) 自动生成路由和服务注册

```shell
go install github.com/yiigo/yiigo/cmd/protoc-gen-yiigo@latest
```

进一步了解 👉 [详情](cmd/protoc-gen-yiigo/README.md)

## gg

受 `protoc-gen-go` 启发，为结构体生成 `Get` 方法【支持泛型!!!】，避免空指针引起的Panic

```shell
go install github.com/yiigo/yiigo/cmd/gg@latest
```

进一步了解 👉 [详情](cmd/gg/README.md)

**Enjoy 😊**
