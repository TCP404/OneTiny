# AGENTS.md instructions

## 基本约定

- 始终使用简体中文回复。
- 运行 shell 命令时默认加 `rtk` 前缀。
- 查找文件和文本时优先使用 `rg` 或 `rg --files`；定位实现位置时优先使用可用的 Semble 搜索。
- 保留用户已有改动。不要执行 `git reset --hard`、`git checkout -- <file>` 等会丢弃改动的命令，除非用户明确要求。

## 分支与 Worktree 规则

- OneTiny 是单仓库项目。代码、文档、配置和构建脚本改动默认在项目内 `.worktrees/` 下的独立 worktree 中完成。
- 如果用户指定复用已有 worktree，先检查分支和工作区状态，再继续工作；不要重复创建同一任务的 worktree。
- 分支命名优先使用 `boii/<short-topic>`，例如 `boii/update-agents-doc`。
- worktree 目录命名优先使用 `.worktrees/<short-topic>`，目录名不要包含 `/`。
- 新建 worktree 前确认 `.worktrees/` 已被 `.gitignore` 忽略；不要在 `.worktrees/` 内再创建嵌套 worktree。

示例：

```bash
rtk git fetch origin
rtk git worktree add .worktrees/update-agents-doc -b boii/update-agents-doc
```

## 开始修改前

- 进入 worktree 后先执行 `rtk git status --short --branch`，识别已有改动并确认本次任务边界。
- 改代码前阅读 [docs/architecture.md](docs/architecture.md) 和 [docs/development.md](docs/development.md)；涉及构建或 CI 时额外阅读 [docs/build.md](docs/build.md)。
- 涉及配置、运行态、CLI/GUI 边界、server 生命周期、生成物或发版流程时，先对照对应文档确认职责边界。

## 代码规范

### Go 与架构边界

- Go 格式化、源码迁移修复和 lint 统一通过 `rtk task precommit` 执行；不要手工维护一套分散命令。
- 新增导出的类型、函数、常量或方法前先做可见性检查；没有跨包真实调用方时保持未导出。
- 不要新增 `pkg/`。当前项目没有对外 Go API 承诺，共享代码优先放在职责明确的 `internal/...` 包中。
- 不要恢复 `internal/conf`、`internal/runtimeconf`、`internal/control`、`internal/handle`、`internal/constant`。
- 配置读写只能通过 `internal/config.Store`；不要在其他包直接操作 Viper、YAML 文件或配置路径。
- 运行态只能由持久化配置和当前进程信息派生，不要把 `Output`、`OS`、`IP`、`Pwd`、`SessionVal` 等运行时字段写入配置。
- CLI-only 逻辑放 `cmd/cli/`，GUI bootstrap 放 `cmd/gui/`，Wails adapter 放 `internal/gui/`，跨 CLI/GUI 的业务编排放 `internal/app/`。
- 错误处理沿用仓库现有风格；需要构造或格式化错误时优先使用 `github.com/pkg/errors`。
- 测试优先验证行为，不验证实现细节；不要为了测试暴露生产 API。
- 避免无边界的 `util`、`common`、`helper`。新增包前先确认职责、真实调用方和依赖方向。

### 前端与 Wails

- 前端源码在 `frontend/src/`；`frontend/bindings/` 是生成物，不手改临时 binding。
- 改 Go service 暴露面后，重新生成 Wails bindings，并提交生成结果。
- 改前端源码后，运行前端构建并提交 `internal/gui/webassets/` 中应提交的生成结果。
- `internal/gui` 只做 Wails adapter，不放配置写入规则、server 核心状态机、文件共享领域逻辑或密码哈希规则。

### 构建与生成物

- 本地快捷入口放 `justfile`，canonical 构建任务放 `Taskfile.yml`，长期构建规则放 `tools/buildtool`。
- 不要在 GitHub Actions 中重复手写 artifact 命名、zip、checksum、ldflags 或 target 解析规则；这些规则应进入 `tools/buildtool`。
- 不应提交本地构建产物：`build/`、`dist/`、临时压缩包、`cmd/gui/rsrc_windows_*.syso`，除非某次发布流程明确要求。
- 修改构建系统时，至少验证 `rtk go test ./...`、`rtk task info TARGET=<target>`，并按影响范围验证 CLI/GUI dist 或 package 任务。

## 文档规则

- 面向维护者的项目文档使用简体中文。
- 架构边界写入 `docs/architecture.md`；开发、测试、生成物和发版约定写入 `docs/development.md`；构建入口、artifact、CI 和 release contract 写入 `docs/build.md`。
- 修改配置、运行态、目录职责、生成物、构建入口或发版流程时，同步更新对应文档。
- 不要把短期任务过程、聊天记录或临时计划写进长期文档。
- README 面向用户，`docs/` 面向维护者；不要把内部架构细节塞进 README，除非它影响用户安装、使用或排障。

## 提交与协作流程

- 不要主动提交或推送，除非用户明确要求。
- 任何提交都必须使用 commit skill，并使用 Conventional Commits。
- 每次提交前必须执行 `rtk task precommit`。该任务会运行 `go fix ./...`、`gofmt` 和 `go vet ./...`；如果它产生文件变更，先检查并按本次任务边界决定是否纳入提交。
- 提交前确认 staged 文件只包含本次任务相关改动；不要把用户的无关改动、构建产物、临时文件或本地配置一起提交。
- 每次推送前必须执行 `rtk task prepush`，其中至少包含 `go test ./...`。
- 按改动类型运行额外验证命令，并在最终说明里列出实际执行过的命令。
- 创建 PR 时使用 `.github/PULL_REQUEST_TEMPLATE.md`，并保留 Summary、Architecture / Config Impact、Test Plan 和 Checklist。
- 默认使用 squash merge，除非维护者明确要求其他方式。
- 不要添加 AI 署名，例如 `Co-Authored-By`、`Generated with` 等。
- 推送前先确认当前分支、远端和用户意图；不要在用户只要求本地提交时推送。

## 发版

OneTiny 走 `release-please`。正常发版不要手动建 tag 或 `gh release create`，把 release-please 创建的 release PR 合并即可触发发版和构建。

示例：

```bash
rtk gh pr list --state open --search "release-please in:title"
rtk gh pr merge <PR_NUMBER> --squash --delete-branch
```

如果 release 被创建为 draft，等构建产物上传完成后再发布 draft。
