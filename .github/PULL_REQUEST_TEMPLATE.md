## Summary

<!-- 用 1-3 条 bullet 说明这次改了什么。 -->
<!-- - 变更点 1 -->

## Architecture / Config Impact

- [ ] 不涉及架构边界、目录职责或依赖方向变化。
- [ ] 涉及架构边界、目录职责或依赖方向变化，已在说明中解释原因，并同步更新 `docs/architecture.md` 或 `docs/development.md`。
- [ ] 不涉及配置字段、配置读写或运行态变化。
- [ ] 涉及配置字段、配置读写或运行态变化，已说明字段属于持久化配置、运行态，还是选择性持久化。

## Test Plan

- [ ] `rtk go test ./...`
- [ ] `rtk go build ./cmd/cli ./cmd/gui`
- [ ] `rtk proxy git diff --check`
- [ ] 前端或 Wails 暴露面变更时：`rtk npm run build --prefix frontend`
- [ ] 未运行某项检查时，已在下方说明原因。

未运行检查说明：

<!-- 如果有检查没跑，在这里说明原因；没有则写“无”。 -->
<!-- - 无 -->

## Checklist

- [ ] PR 标题使用 Conventional Commits。
- [ ] 没有手改 `frontend/bindings/` 或 `internal/gui/webassets/dist/` 这类生成物；如有生成物变更，来源命令已写入 Test Plan。
- [ ] 没有绕过 `internal/config.Store` 写配置。
- [ ] 没有把运行态字段写入 `config.Config`。
- [ ] 没有恢复 `pkg/`、`internal/conf`、`internal/runtimeconf`、`internal/control`、`internal/handle`、`internal/constant`。
- [ ] 没有手改 release-please 维护的 `CHANGELOG.md`；如需改 changelog 规则，改的是 `release-please-config.json`。
