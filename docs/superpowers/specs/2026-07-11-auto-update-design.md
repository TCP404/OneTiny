# OneTiny 自动更新设计

## 背景

OneTiny 同时提供 CLI 和 GUI。当前已有 `onetiny update` 命令，但它仍按旧 release 资产名下载文件，没有替换当前程序，也没有 GUI 入口。Release workflow 里部分版本注入仍指向旧的 `internal/constant.VERSION`，会导致 release 二进制无法正确报告当前版本。

本设计目标是提供 GUI 和 CLI 共用的自动更新能力：检查新版、下载校验、替换当前程序，并在 GUI 启动后默认自动检查一次。下载和安装必须由用户明确触发，第一版不做后台自动下载。

## 目标

- CLI 和 GUI 使用同一套更新核心逻辑，入口层只处理交互差异。
- 支持 `latest` 和指定 tag 更新。
- 支持 release asset 自动匹配当前入口、操作系统和架构。
- 下载 zip 后使用 release checksums 做 sha256 校验。
- 安装时通过独立 helper 等待当前进程退出，再替换文件并尽量重启。
- 修正 Makefile、justfile 和 GitHub release workflow 的版本注入与资产契约。

## 非目标

- 不引入后台静默安装。
- 不新增自动下载或自动检查的持久化配置。
- 不实现系统级提权安装。遇到无写权限目录时返回明确错误，并保留已下载包路径。
- 不改变 release-please 的发版模型，不手动创建 tag。

## 架构

新增 `internal/updater` 作为共享更新领域包。它不依赖 `internal/config`、`internal/runtime`、`internal/app`、`internal/gui` 或 `internal/server`，只负责 release 元数据、版本比较、资产选择、下载校验、解压、安装计划和 helper 执行。

入口分工：

- `cmd/cli/update.go` 只解析 CLI flags、询问确认、输出进度和错误。
- `internal/app.Service` 暴露 GUI 所需的检查、下载、安装用例 DTO。
- `internal/gui.Service` 只做 Wails adapter，把调用转发给 `internal/app.Service`。
- `frontend/src/` 展示更新状态，并在 GUI 启动后调用一次检查接口。

配置影响：

- 第一版不新增 `config.Config` 字段。
- GUI 自动检查是当前 GUI 进程内行为，默认每次启动检查一次。
- 下载中的临时文件和 helper 状态属于当前进程/临时目录状态，不写入配置。

## Release 资产契约

Updater 根据入口类型和平台匹配 zip asset：

- CLI：`onetiny-cli-<os>-<arch>.zip`
- GUI：`onetiny-gui-<os>-<arch>.zip`

其中 `<os>` 使用 `linux`、`windows`、`darwin`，`<arch>` 使用 `x64`、`arm64`。现有 release workflow 已生成这些 GUI/CLI 资产名，设计要求 Makefile、justfile 和 workflow 保持一致。

zip 内容约定：

- CLI zip 内包含单个可执行文件：`onetiny-cli` 或 `onetiny-cli.exe`。
- Windows GUI zip 内包含单个 `OneTiny.exe`。
- macOS GUI zip 内包含 `OneTiny.app` bundle。

Release 必须包含 `onetiny-checksums.txt`，格式沿用 `sha256sum *.zip`。安装前必须验证目标 zip 的 sha256。校验失败时不解压、不替换。

## 版本判断

`internal/updater` 使用语义版本比较：

- 接受 `v0.11.0` 和 `0.11.0`。
- 忽略 GitHub tag 前缀 `v`。
- 当前版本为空或无法解析时显示为 `dev`，GUI 自动检查不提示“可更新”；手动检查可以显示最新 release 信息，但安装前会提示当前构建版本未知。

Release workflow、Makefile 和 justfile 必须统一通过：

```text
-X github.com/tcp404/OneTiny/internal/version.Version=<tag>
```

注入版本。旧的 `internal/constant.VERSION` 注入必须移除。

## CLI 行为

`onetiny update` 默认检查 latest，如果有新版则询问是否下载并安装。

Flags：

- `--check`：只检查新版，不下载。
- `--download-only`：下载并校验 zip，输出保存路径，不安装。
- `--yes`：跳过确认，适合脚本。
- `--use <tag>`：安装指定版本。
- `--list`：列出远端 tag，保留现有语义。
- `--output <dir>`：指定 download-only 保存目录；未指定时使用用户下载目录或临时目录。

CLI 安装流程：

1. 查询 release。
2. 匹配 CLI asset。
3. 下载 zip 和 checksums。
4. 校验 sha256。
5. 解压到临时 staging 目录。
6. 复制当前可执行文件到临时 helper 路径。
7. 启动 helper，传入当前 PID、当前 exe 路径、新 exe 路径、备份路径和可选重启参数。
8. 当前 CLI 退出，让 helper 完成替换。

默认 CLI 更新完成后不自动重新启动服务进程；用户重新运行 `onetiny` 即可。若未来需要 `--restart`，可以复用 helper 的重启参数。

## GUI 行为

GUI 启动后，前端在首次渲染完成后调用一次 `CheckUpdate`。检查失败不阻塞启动，失败信息只在“关于”页或用户手动检查时展示，避免离线环境反复打扰。

GUI 新增“关于”页更新区域：

- 当前版本。
- 最新版本。
- 状态：检查中、已是最新、有新版本、下载中、已下载、安装失败。
- 操作：检查更新、下载更新、安装并重启。

GUI 下载和安装流程：

1. 用户点击下载。
2. `internal/app.Service` 调用 `internal/updater` 下载、校验并解压到 staging。
3. 前端显示“已下载，可安装”。
4. 用户点击安装并重启。
5. GUI 提示安装会退出应用并停止共享服务。
6. 用户确认后，GUI 调用安装接口启动 helper。
7. GUI 调用现有 shutdown 路径停止 sharing，然后 `app.Quit()`。
8. helper 等旧进程退出，替换文件，再尝试重启 GUI。

## 安装与重启策略

当前进程不直接覆盖自己。所有平台都使用临时 helper，行为一致：

- helper 是当前可执行文件复制到临时目录后的副本，使用隐藏内部参数进入 apply-update 模式。
- helper 等待原 PID 退出，超时后返回错误。
- helper 替换前创建 `.bak` 备份。
- 替换成功后删除备份；替换失败时尽量回滚。
- helper 日志写到临时目录，GUI/CLI 返回 helper 启动结果和日志路径。

平台差异：

- Windows CLI/GUI：必须等待旧进程退出后替换 `.exe`，因为运行中的 `.exe` 通常无法覆盖。
- macOS GUI：优先替换整个 `OneTiny.app` bundle，然后用 `open <OneTiny.app>` 重启。若当前不是从 `.app` 运行，则退化为替换当前二进制。
- Linux GUI：当前 release 不提供 GUI 资产，若没有匹配 asset，GUI 展示“不支持当前平台自动更新”。CLI 可按资产存在情况更新。
- Unix CLI：即使技术上可 rename 运行中的可执行文件，也仍使用 helper，保持跨平台一致。

权限不足时不自动提权。错误消息包含目标路径、原因和已下载包路径，方便用户手动替换。

## 数据模型

`internal/updater` 核心类型：

- `Channel`：`cli` 或 `gui`。
- `Release`：tag、name、body、published time、assets。
- `Asset`：name、download URL、size、content type。
- `CheckResult`：current version、latest version、available、release URL、matched asset、reason。
- `DownloadResult`：release、asset、zip path、checksum、staging path、install candidate path。
- `InstallPlan`：current PID、current path、replacement path、backup path、restart command、log path。

GUI DTO 由 `internal/app` 定义，字段保持前端友好，不直接暴露 updater 内部结构。

## 错误处理

错误分为可展示的明确原因：

- 网络不可用或 GitHub API 返回错误。
- 当前版本未知。
- 没有匹配当前平台的 release asset。
- 缺少 checksums 文件。
- sha256 校验失败。
- zip 结构不符合契约。
- staging 解压失败。
- helper 启动失败。
- 目标目录无写权限。
- 替换失败并回滚。

自动检查失败不弹出强提示。手动检查、下载、安装失败必须展示错误。

## 测试计划

Go 单元测试：

- 语义版本比较：`v` 前缀、相等、升级、降级、预期 dev 行为。
- asset 匹配：CLI/GUI、OS、arch、缺失资产。
- checksums 解析和 sha256 校验。
- zip 解压结构验证。
- install plan 路径生成，包含 Windows exe、macOS app bundle、普通二进制。
- helper apply 逻辑使用临时目录模拟替换和回滚。

入口构建验证：

```bash
rtk go test ./...
rtk go build ./cmd/cli ./cmd/gui
rtk npm run build --prefix frontend
rtk proxy git diff --check
```

发版配置验证：

```bash
rtk grep -n "internal/constant\\.VERSION" .github cmd internal justfile Makefile Taskfile.yml
rtk grep -n "internal/version\\.Version" .github justfile Makefile
```

手工验证：

- 用本地临时 release fixture 测试 CLI `--check`、`--download-only`、`--use`。
- GUI 浏览器预览模式验证更新区域布局不溢出。
- Windows 上验证 GUI 退出后 helper 能替换 `.exe`。
- macOS 上验证 `.app` 替换后能通过 `open` 重启。

## 实施顺序

1. 修正 release workflow、Makefile、justfile 的版本注入和资产契约。
2. 新增 `internal/updater` 的查询、版本、资产和校验能力。
3. 实现 download-only 和 staging 解压。
4. 实现 helper apply-update 模式和安装计划。
5. 重写 CLI `update` 命令使用共享 updater。
6. 在 `internal/app`、`internal/gui` 和 frontend 增加 GUI 更新入口。
7. 补测试并跑架构回归检查。
