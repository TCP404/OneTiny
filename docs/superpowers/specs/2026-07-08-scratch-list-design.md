# 临时列表设计

## 背景

OneTiny 当前以 `RootPath` 下的文件和目录为共享对象。用户在不同机器之间临时传递一段文字或图片时，不一定想先在共享目录里创建文件。需要新增一个进程内的临时共享列表，方便局域网设备直接粘贴和读取内容；程序退出后内容消失。

## 目标

- 支持临时共享多条文本和图片。
- 内容只保存在当前进程内存中，不写入 `RootPath`，不落盘。
- 远程设备可以新增临时内容。
- 不提供手动删除。
- 按最近提交或重复命中的顺序显示，最新在上。
- 重复内容不新增，而是把已有条目提到顶部。
- 列表容量和单条大小上限可配置。
- 开启 `IsSecure` 时复用现有登录保护。
- CLI 支持启动 flag 和 `config` 子命令配置；GUI 设置页支持修改配置。
- CLI 不提供终端内粘贴命令，用户通过浏览器访问临时列表页面粘贴。

## 非目标

- 不支持目录或任意小文件的临时共享。
- 不支持手动删除、按用户删除或创建者权限。
- 不支持跨进程持久化、重启恢复或多实例同步。
- 不新增独立权限系统。
- 不改变 `/` 默认跳转 `/file/` 的行为。

## 推荐方案

新增和 `/file/` 并列的 `/scratch/` 页面与接口。

这个方案符合当前服务端共享页的组织方式：`resource/template` 使用服务端模板，Gin route 负责页面和提交接口，登录保护通过 middleware 统一处理。它避免把临时内容混进 RootPath 文件浏览，也避免第一版只提供 API 而缺少可用入口。

## 架构

新增 `internal/scratch/` 包，职责是内存列表领域逻辑：

- item 模型。
- 内容 hash。
- 去重提顶。
- 容量淘汰。
- 单条大小校验。
- 并发安全。

新增 `internal/server/handler/scratch/` 包，职责是 HTTP 层：

- 渲染临时列表页面。
- 接收文本和图片提交。
- 输出 item 原始内容。
- 将领域错误转换为 HTTP 响应。

新增 `internal/server/routes/scratch.go`：

- 注册 `/scratch` 路由组。
- 对所有 `/scratch/*` 路由复用 `middleware.CheckLogin`。
- 不使用 `middleware.CheckLevel`，因为临时列表不属于 RootPath 文件树。

`server.Manager` 持有一个 `scratch.Store`，并在每次构建 Gin engine 时注入给 routes。这样端口重启重建 HTTP server 时临时列表不会丢失；进程退出后 manager 和 store 一起释放。

配置读写仍只通过 `internal/config.Store`。临时内容本身不是配置，只把容量和大小上限作为配置保存。

## 数据模型

`internal/scratch.Item`：

```go
type Item struct {
    ID        string
    Kind      Kind
    MimeType  string
    Data      []byte
    Size      int64
    Hash      string
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

`Kind` 第一版只支持：

- `text`
- `image`

文本保存 UTF-8 bytes，MIME 使用 `text/plain; charset=utf-8`。图片保存浏览器提交的原始 bytes，第一版支持：

- `image/png`
- `image/jpeg`
- `image/gif`
- `image/webp`

hash 使用 `sha256(kind + mimeType + data)`。这样可以避免不同类型但 bytes 相同的内容被错误合并。

## 排序、去重和淘汰

列表按最近提交或重复命中排序，最新在上。

新增内容时：

1. 校验 kind、MIME 和大小。
2. 计算 hash。
3. 如果 hash 已存在，不新增 item，把已有 item 提到顶部，并刷新 `UpdatedAt`。
4. 如果 hash 不存在，创建 item 并放到顶部。
5. 如果总数超过容量，从底部淘汰最旧 item。

示例：依次提交 `1 2 3 4 5 4 2`，页面从上到下显示 `2 4 5 3 1`。

不提供手动删除。清理只通过容量淘汰和进程退出完成。

## HTTP 路由

新增路由：

- `GET /scratch/`：渲染临时列表页面。
- `POST /scratch/items`：新增文本或图片。
- `GET /scratch/items/:id`：返回原始内容。

`POST /scratch/items` 支持 multipart/form-data：

- 文本字段：`kind=text`，`text=<内容>`。
- 图片字段：`kind=image`，`image=<文件>`。

成功响应规则：

- 页面粘贴交互使用 fetch 提交，返回 JSON。
- 普通表单提交成功后重定向回 `/scratch/`。
- 提交失败返回明确错误；fetch 场景返回 JSON 错误，普通表单场景返回错误页面或带错误信息的 `/scratch/` 页面。

`GET /scratch/items/:id` 根据 item 设置响应：

- `Content-Type` 使用 item 的 MIME。
- 图片可以浏览器预览。
- 文本可以直接返回原文。
- 下载链接可设置 `Content-Disposition: attachment`。

## 页面设计

新增 `resource/template/scratch.tpl`。

页面能力：

- 顶部提供 `/file/` 与 `/scratch/` 切换入口。
- 文本区域用于输入文本并提交。
- 图片选择按钮用于选择图片提交。
- 页面监听 `paste` 事件：
  - 粘贴文本时提交为 text item。
  - 粘贴图片时提交为 image item。
- 列表展示类型、大小和更新时间。
- 文本 item 展示预览和复制按钮。
- 图片 item 展示缩略图、预览和下载入口。
- 无删除按钮。

`resource/template/list.tpl` 增加临时列表入口。`/` 默认仍跳转 `/file/`。

## 安全

`/scratch` 路由组复用 `middleware.CheckLogin`。

- `IsSecure=false`：局域网访问者可以查看和新增临时内容。
- `IsSecure=true`：未登录先进入 `/login`，登录后才能访问 `/scratch/*`。

不新增“创建者”或“删除者”身份。第一版不提供删除功能，因此不需要按用户区分权限。

## 配置

新增持久化配置项：

```yaml
scratch:
  max_items: 500
  max_item_size: 10MB
```

配置默认值：

- `ScratchMaxItems = 500`
- `ScratchMaxItemSize = "10MB"`

运行态派生字段：

- `ScratchMaxItems int`
- `ScratchMaxItemSize string`
- `ScratchMaxItemBytes int64`

`scratch.max_item_size` 支持大小字符串，例如 `10MB`。配置读取后保留原始字符串用于 GUI 和 CLI 展示，同时解析为 bytes 放入运行态供服务端校验。解析失败应让配置加载或配置更新失败，而不是在新增 item 时才失败。

配置改动需要更新：

- `internal/defaults`
- `internal/config.Config`
- `internal/config.ConfigPatch`
- `configSettings`
- `configFromViper`
- `internal/runtime.PersistentConfig`
- `runtime.Snapshot`
- `runtime.Patch`
- CLI flag 和 config 命令
- GUI DTO 和设置页
- `docs/architecture.md` 的配置分类

## 校验

配置校验规则：

- `ScratchMaxItems >= 1`
- `ScratchMaxItemBytes >= 1`
- `ScratchMaxItemSize` 必须能解析为正数 bytes。

新增 item 校验规则：

- 文本不能为空。
- 图片 MIME 必须是支持列表之一。
- 单条内容大小不能超过 `ScratchMaxItemBytes`。
- 任一校验失败时不改变现有列表。

错误文案应明确，例如：

- `临时内容不能为空`
- `不支持的临时内容类型`
- `临时内容超过 10MB 上限`
- `临时列表容量必须大于 0`

## CLI

新增 root flag：

```bash
onetiny --scratch-max-items 500 --scratch-max-item-size 10MB
```

root flag 只覆盖本次启动的运行态，不写配置。

新增 config 子命令支持：

```bash
onetiny config --scratch-max-items 500 --scratch-max-item-size 10MB
```

config 子命令写入配置文件后退出。

启动信息增加临时列表地址，例如：

```text
Run on        [ http://192.168.1.10:8192 ]
Scratch list [ http://192.168.1.10:8192/scratch/ ]
```

不新增 `onetiny scratch add`。终端内新增内容会涉及运行中进程发现、HTTP 调用和登录 session，第一版通过浏览器页面粘贴更符合使用场景。

## GUI

GUI 设置页增加两个配置项：

- 临时列表容量。
- 单条大小上限。

保存时走 `internal/app.Service.UpdateConfig`：

- 写入 `internal/config.Store`。
- 派生新的 runtime snapshot。
- 调用 `server.Manager.ApplyRuntime` 更新运行态。
- 不需要重启 HTTP server。

Wails binding 需要重新生成，`frontend/bindings/` 不手改。

## 测试计划

`internal/scratch`：

- 新增文本。
- 新增图片。
- 重复 hash 提到顶部。
- 容量淘汰底部 item。
- 单条大小超限拒绝。
- 不支持 MIME 拒绝。
- 并发新增不产生数据竞争。

`internal/config`：

- 默认配置包含 scratch 默认值。
- 读取旧配置时回填 scratch 默认值。
- patch 后写入并重新 load。
- 无效容量和无效大小上限被拒绝。

`internal/runtime`：

- 配置派生到 snapshot。
- runtime patch 能更新 scratch 限制。

CLI：

- root flag 覆盖本次运行态。
- `config` 子命令写入 scratch 配置。
- 启动信息包含 `/scratch/` 地址。

HTTP：

- `/scratch/` 在安全模式下受登录保护。
- 文本提交成功后列表展示。
- 图片提交成功后可以读取原始内容。
- 超限提交返回错误且列表不变。
- 重复提交不新增条目并提到顶部。

模板：

- `scratch.tpl` 包含粘贴、文本输入、图片选择和列表入口。
- `list.tpl` 包含跳转到临时列表的入口。

构建：

```bash
rtk go test ./...
rtk go build ./cmd/cli ./cmd/gui
rtk npm run build --prefix frontend
rtk proxy git diff --check
```

## 迁移和兼容

旧配置文件没有 `scratch` 节时使用默认值，不需要用户手动迁移。

新增配置字段属于持久化配置，但临时 item 内容不属于配置。程序退出后 item 丢失是预期行为。

## 设计结论

第一版实现一个独立 `/scratch/` 临时列表页面，用内存 store 保存文本和图片。列表容量和单条大小上限可配置，CLI/GUI 都能修改限制。内容不落盘、不进入 RootPath、不支持手动删除，并复用现有登录保护。
