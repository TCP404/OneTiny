# OneTiny 构建系统

本文档定义 OneTiny 的长期构建契约。目标是让本地开发和 CI 使用同一套入口，避免构建规则散落在 `justfile`、`Taskfile.yml` 和 GitHub Actions 脚本里。

## 分层职责

构建系统分三层：

- `justfile`：本地体验层。只保留短命令、别名和开发者常用入口，不承载真实构建规则。
- `Taskfile.yml`：构建编排层。本地和 CI 的 canonical 入口，负责串联 check、build、package、dist 任务图。
- `tools/buildtool`：构建规则层。固化 target 解析、artifact 命名、输出路径、zip、checksum、版本校验和产物校验。

GitHub Actions 只负责准备平台环境、调用 `task`、上传产物或 release assets。不要在 workflow 里重复手写 `go build`、`npm run build`、`wails3 generate syso`、压缩包命名或 checksum 规则。

## Target 和 Artifact

构建目标使用 `GOOS-GOARCH` 格式：

- `linux-amd64`
- `windows-amd64`
- `darwin-amd64`
- `darwin-arm64`

发布 artifact 使用面向用户的 label：

- `linux-amd64` -> `linux-x64`
- `windows-amd64` -> `windows-x64`
- `darwin-amd64` -> `darwin-x64`
- `darwin-arm64` -> `darwin-arm64`

CLI artifact 命名：

```text
onetiny-cli-<platform-label>.zip
```

GUI artifact 命名：

```text
onetiny-gui-<platform-label>.zip
```

checksum 文件固定为：

```text
dist/onetiny-checksums.txt
```

## 常用命令

本地优先使用 `just`：

```bash
rtk just info
rtk just check
rtk just cli
rtk just gui
rtk just dev
rtk just hooks-install
rtk just precommit
rtk just prepush
```

需要复现 CI 时使用 `task`：

```bash
rtk task info TARGET=windows-amd64
rtk task dist:cli TARGET=linux-amd64 VERSION=v0.6.1
rtk task dist:gui TARGET=windows-amd64 VERSION=v0.6.1
rtk task package:gui TARGET=darwin-arm64 VERSION=v0.6.1
```

## 当前平台边界

CLI 可以在 Linux runner 上交叉编译多个平台目标。

GUI 应在目标系统 runner 上构建：

- macOS `.app` 和 codesign 在 macOS runner 上执行。
- Windows GUI 和 `.syso` 在 Windows runner 上执行。
- Linux GUI/package 在 Linux runner 上执行。

`tools/buildtool` 可以集中规则，但不会改变底层平台限制。

## 扩展规则

新增构建规则时按以下顺序放置：

1. 影响 artifact 名称、target、路径、zip、checksum、版本或产物校验的规则，放进 `tools/buildtool` 并加 Go 测试。
2. 只负责串任务的流程，放进 `Taskfile.yml`。
3. 只为本地输入更短或更顺手的入口，放进 `justfile`。
4. CI workflow 只能增加环境准备、缓存、artifact 上传和 release 上传步骤。

Windows installer、Linux AppImage/deb/rpm、macOS notarization、签名和发布前校验都应优先扩展 `tools/buildtool`，再由 `Taskfile.yml` 调用。

## 生成物

这些是构建产物，不应作为源文件维护：

- `build/`
- `dist/`
- `cmd/gui/rsrc_windows_*.syso`

这些生成物根据项目开发规范提交：

- `frontend/bindings/`
- `internal/gui/webassets/`

修改前端或 Wails 暴露面后，仍需按 `docs/development.md` 的规则生成并提交对应文件。
