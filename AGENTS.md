# AGENTS.md instructions

- 始终使用简体中文回复。
- 运行 shell 命令时默认加 `rtk` 前缀。
- 改代码前先阅读 `docs/architecture.md` 和 `docs/development.md`，尤其是配置可信源、运行态和目录职责。
- 不要恢复 `pkg/`、`internal/conf`、`internal/runtimeconf`、`internal/control`、`internal/handle`、`internal/constant`。
- 配置读写只能通过 `internal/config.Store`；运行态只能从配置和当前进程派生。
- CLI-only 代码放 `cmd/cli/`，GUI-only bootstrap 放 `cmd/gui/`，Wails adapter 放 `internal/gui/`。
- `frontend/bindings/` 是生成物，不手改临时 binding。
- 默认使用 squash merge，除非维护者明确要求其他方式。

## 发版

OneTiny 走 `release-please`。正常发版不要手动建 tag 或 `gh release create`，把 release-please 创建的 release PR 合并即可触发发版和构建。

示例：

```bash
rtk gh pr list --state open --search "release-please in:title"
rtk gh pr merge <PR_NUMBER> --squash --delete-branch
```

如果 release 被创建为 draft，等构建产物上传完成后再发布 draft。
