# AGENTS.md instructions

- 始终使用简体中文回复。
- 运行 shell 命令时默认加 `rtk` 前缀。

## 发版

OneTiny 走 `release-please`。正常发版不要手动建 tag 或 `gh release create`，把 release-please 创建的 release PR 合并即可触发发版和构建。

示例：

```bash
rtk gh pr list --state open --search "release-please in:title"
rtk gh pr merge <PR_NUMBER> --merge --delete-branch
```
