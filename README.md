# web

`web` 是 infrago 的**模块**。

## 包定位

- 类型：模块
- 作用：Web 模块，负责站点、多路由、静态资源与业务整合。

## 主要功能

- 对上提供统一模块接口
- 对下通过驱动接口接入具体后端
- 支持按配置切换驱动实现

## 快速接入

```go
import _ "github.com/infrago/web"
```

```toml
[web]
driver = "default"
```

## 驱动实现接口列表

以下接口由驱动实现（来自模块 `driver.go`）：

### Driver

- `Connect(*Instance) (Connection, error)`

### Connection

- `Open() error`
- `Close() error`
- `Register(name string, info Info, hosts []string) error`
- `Start() error`
- `StartTLS(certFile, keyFile string) error`

### Delegate

- `Serve(name string, params Map, res http.ResponseWriter, req *http.Request)`

## 站点跨域配置

跨域只支持站点级配置，不支持全局 `[cross]` / `[web.cross]`。

- `[site.xxx.cross]`
- 或 `cross = { enable = true, ... }`

常用键：

- `allow` / `enable`
- `method` / `methods`
- `origin` / `origins`
- `header` / `headers`

## 全局配置项（所有配置键）

配置段：`[web]`

- `driver`
- `port`
- `host`
- `bind`
- `cert`
- `certfile`
- `key`
- `keyfile`
- `charset`
- `cookie`
- `token`
- `expire`
- `crypto`
- `maxage`
- `httponly`
- `answerencode`
- `codec`
- `answer`
- `upload`
- `static`
- `shared`
- `defaults`
- `domain`
- `domains`
- `alias`
- `aliases`
- `setting`

## 说明

- `setting` 一般用于向具体驱动透传专用参数
- 多实例配置请参考模块源码中的 Config/configure 处理逻辑
