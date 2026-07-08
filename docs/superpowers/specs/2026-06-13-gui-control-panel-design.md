# OneTiny GUI 控制面板设计

## 背景

OneTiny 当前是一个 CLI 驱动的局域网文件共享工具。程序启动后运行 Gin HTTP server，通过内嵌的 `login.tpl` 和 `list.tpl` 给局域网客户端渲染登录页和文件列表页，并把用户配置保存到 `{UserConfigDir}/tiny/config.yml`。

本设计增加一个跨平台桌面 GUI 控制面板。第一版只做本机控制面板，不改局域网客户端访问时看到的文件列表页。

## 目标

- 提供 Windows、macOS、Linux 三端桌面 GUI。
- 使用 Go 后端，加 Wails v3 承载 HTML/CSS/JS 桌面面板。
- 打开 GUI 后不自动共享，用户点击“启动共享”后才启动服务。
- 运行中支持热更新共享目录、上传开关、登录开关和最大访问层级。
- 运行中修改端口必须先弹确认，确认后重启 HTTP 服务。
- 点击窗口关闭按钮时隐藏到托盘，不退出程序。
- 托盘左键打开控制面板。
- 托盘右键菜单包含“打开面板”和“退出”。
- 访问日志保存到本地日志文件。
- GUI、CLI 和 Web 登录统一从 MD5 账号密码存储迁移到 bcrypt。

## 非目标

- 第一版不重做局域网客户端文件列表页。
- 第一版不提供局域网远程管理后台。
- 第一版不做临时访问码、按客户端授权等高级权限能力。

## 推荐架构

实现时应把当前全局启动流程拆成四个主要服务。

### ServiceManager

负责 Gin HTTP server 的生命周期。

- `Start(config)`: 启动文件共享服务。
- `Stop()`: 优雅停止服务。
- `Restart(config)`: 重启服务，主要用于端口变化。
- `ApplyRuntimeConfig(config)`: 热更新非端口运行时配置。
- `Status()`: 返回运行状态、访问地址和当前配置。

热更新规则：

- 共享目录：下一次请求立即使用新目录。
- 上传开关：上传 handler 立即生效。
- 登录开关：登录 middleware 立即生效。
- 最大访问层级：层级校验 middleware 立即生效。
- 端口：必须用户确认后重启服务。
- 账号密码：保存后立即影响新的登录验证。

### ConfigStore

负责读取和写入 `{UserConfigDir}/tiny/config.yml`。

它应成为 CLI 和 GUI 共用的唯一配置来源，避免出现两套行为。

推荐配置结构：

```yaml
server:
  road: /Users/me/Downloads
  port: 8192
  allow_upload: false
  max_level: 1

account:
  secure: false
  custom:
    user: admin
    pass_hash: "$2a$10$..."
    pass_hash_algo: bcrypt
```

如果检测到旧字段 `account.custom.pass`，应把它视为不兼容的旧安全配置。若此时开启了登录保护，服务启动必须失败，并给出明确提示，要求用户重新设置账号密码。

### CredentialService

负责账号设置和登录验证。

- `SetCredentials(username, password)`: 写入 bcrypt hash。
- `Verify(username, password)`: 验证 Web 登录。
- `IsConfigured()`: 判断是否存在可用账号密码。
- `ValidateCredentialConfig()`: 拒绝旧 MD5-only 配置。

CLI `sec` 命令、GUI 账号弹窗和 Web 登录 handler 都必须使用这同一套服务。

### AccessLogger

负责把访问日志写入 `{UserConfigDir}/tiny/access.log`，格式使用 JSON Lines。

示例：

```json
{"time":"2026-06-13T18:42:18+08:00","client_ip":"192.168.1.23","method":"GET","event":"download","path":"/release/OneTiny.exe","status":200,"result":"success"}
```

需要记录的事件：

- 文件列表访问。
- 文件下载。
- 上传成功和上传被拒绝。
- 登录成功和登录失败。
- 访问被拒绝。
- 与客户端请求相关的服务端错误。

GUI 需要支持读取、筛选、清空和导出日志。

## 桌面应用

桌面壳建议使用 Wails v3，因为当前需求可以接受少量前端工具链，并且需要跨平台窗口、托盘、菜单和原生目录选择器能力。

Wails 后端建议暴露给前端的方法：

- `GetStatus()`
- `StartSharing()`
- `StopSharing()`
- `ChooseDirectory()`
- `UpdateConfig(patch)`
- `SetCredentials(username, password)`
- `GetLogs(filter)`
- `ClearLogs()`
- `ExportLogs()`
- `OpenConfigDir()`

## 页面

### 控制面板

顶部区域是主控制面：

- `访问地址` 标题。
- 运行状态 label 放在标题后面：`未运行` 或 `运行中`。
- 访问地址展示。
- `复制地址` 按钮。
- `启动共享` / `停止共享` 按钮。

顶部区域下方包含：

- 共享目录选择器。
- 端口输入。
- 最大访问层级输入。
- 上传开关。
- 登录保护开关。
- 账号设置按钮。

行为：

- 打开 GUI 不自动启动共享。
- 点击“启动共享”才启动 Gin 服务。
- 只有服务运行后才显示访问地址。
- 共享目录、上传、登录和最大访问层级在运行中热更新。
- 运行中修改端口时弹确认，确认后重启服务。

### 安全设置

集中管理登录保护和账号密码。

- 登录保护开关与控制面板里的开关同步。
- “修改账号密码”打开与控制面板相同的账号设置弹窗。
- 未配置账号密码时，点击登录开关会打开账号设置弹窗。
- 保存有效的账号、密码、确认密码后写入 bcrypt 配置，并自动开启登录保护。

未来可选安全设置可以放在这里：

- 会话过期时间。
- 登录失败处理。
- 客户端 IP 限制。

这些不是第一版必需能力，除非实现成本很低。

### 访问日志

从 `{UserConfigDir}/tiny/access.log` 展示持久化访问日志。

表格列：

- 时间。
- 客户端 IP。
- 事件。
- 路径。
- 结果。

控件：

- 按事件类型筛选。
- 按时间范围筛选。
- 清空日志。
- 导出日志，第一版建议导出 CSV。

日志文件不存在或为空时，页面应正常展示空状态。

### 关于

展示：

- 应用版本。
- 项目说明。
- 配置文件路径。
- 访问日志文件路径。
- 打开配置目录按钮。

## 托盘行为

- 点击窗口关闭按钮时隐藏到托盘，不退出程序。
- 托盘左键显示并聚焦主窗口。
- 托盘右键菜单包含：
  - 打开面板。
  - 退出。
- 服务运行中点击退出时，先弹确认，再停止服务并退出。

## CLI 与迁移行为

CLI 继续支持。

- `sec` 必须写入 bcrypt 凭证。
- Web 登录必须验证 bcrypt 凭证。
- GUI 账号设置必须写入同一份 bcrypt 配置。
- 旧 MD5 凭证不能静默迁移，因为无法从 MD5 还原原密码。
- 如果开启登录保护时检测到旧 MD5 凭证，服务启动失败，用户必须重新设置账号密码。

## 测试范围

### 服务生命周期

- 停止状态下启动成功。
- 运行状态下停止成功。
- 确认修改端口后重启成功。
- 重启失败时有明确状态和错误提示。

### 运行时热更新

- 共享目录变化影响后续请求。
- 上传开关立即影响上传 handler。
- 登录开关立即影响受保护文件路由。
- 最大访问层级立即影响目录访问。
- 端口变化必须确认并重启服务。

### 账号密码

- 新账号密码以 bcrypt 存储。
- Web 登录使用正确账号密码成功。
- Web 登录使用错误账号密码失败。
- CLI `sec` 写入 bcrypt 配置。
- 旧 MD5 配置阻止受保护服务启动，并给出明确错误。

### 访问日志

- 访问、下载、上传、登录成功/失败、访问被拒绝和服务端错误都会记录。
- GUI 可以读取、筛选、清空和导出日志。
- 日志文件不存在时展示空状态。

### 桌面行为

- 关闭窗口隐藏到托盘。
- 托盘左键打开面板。
- 托盘右键菜单动作可用。
- 服务运行中退出时出现确认。

## 风险

- Wails v3 仍是 alpha，API 可能变化。
- Linux 托盘行为会受桌面环境影响。
- bcrypt 迁移对旧 MD5 登录保护配置是刻意的破坏性变更。
- 当前全局 `config.Config` 需要谨慎重构，避免运行时并发读写不安全。

## 实施顺序

1. 抽出 `ConfigStore`、`CredentialService`、`ServiceManager` 和 `AccessLogger`。
2. 把 CLI `sec` 和 Web 登录改成 bcrypt。
3. 给 Gin 服务增加启动、停止、重启和运行时配置更新能力。
4. 增加访问日志 middleware 和日志读取/导出能力。
5. 搭建 Wails v3 桌面壳、托盘和原生目录选择器。
6. 实现 HTML/CSS/JS GUI 页面。
7. 更新 README，补充 GUI 使用说明和 MD5 迁移说明。
8. 跑单元测试、集成测试，并在 Windows、macOS、Linux 上做手动 QA。
