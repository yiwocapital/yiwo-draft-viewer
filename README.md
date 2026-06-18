# YiwoDraftViewer

macOS Markdown 稿件预览工具。Go + Wails v2 + 原生前端（无框架），专为"渐进式撰稿"工作流设计。

## 特性

- **单文件 Markdown 预览** — 拖拽或菜单打开 `.md` 文件，自动识别 YAML frontmatter
- **Git 历史浏览** — 自动发现文件所在仓库，按时间倒序显示所有 commit
- **字级 diff 高亮** — 内置 `diff-match-patch`，绿色显示新增、红色删除线显示删除
- **多选对比** — Cmd+点击两个 commit（或先单选与最新对比，再 Cmd+点击锁定两个）
- **字数统计** — 折叠空白计算，HTML 编辑注释自动剔除
- **编辑注释折叠** — 切换按钮隐藏/灰显 `<!-- -->` 编辑笔记
- **结构化复制** — 一键复制标题/摘要/正文/全部；自动修复英文标点为中文（如 `"hello"` → `"hello"`，`!` → `！`）
- **字号可调** — `Cmd++` / `Cmd+-`，配置持久化
- **窗口记忆** — 大小/位置/字号自动恢复
- **通读模式** — 一键切换纯文本浏览（不显示 diff 高亮）
- **状态栏** — 显示当前文件绝对路径，点击复制
- **Toast 通知** — 复制成功自动 2.5s 消失，错误需手动关闭

## 系统要求

- macOS 11+（Wails v2 + WKWebView）
- 无其他依赖（自包含 `.app`）

## 安装

下载对应 tag 的 release（`.app` 或 `.zip`），拖入 `~/Applications/`。

或从源码构建：

```bash
git clone https://github.com/yiwocapital/yiwo-draft-viewer
cd yiwo-draft-viewer
make build
rsync -a --delete build/bin/yiwo-draft-viewer.app/ ~/Applications/yiwo-draft-viewer.app/
```

## 使用

1. 启动 app（从 `~/Applications/` 或 Spotlight 搜索 "YiwoDraftViewer"）
2. **File → Open** 或拖入 `.md` 文件
3. 右侧栏显示 commit 列表，点击切换 diff 视图
4. **Cmd+点击** commit 切换对比模式（橙色高亮）
5. **📖 通读全文** 按钮（或按 `Esc`）进入纯文本模式
6. **Cmd++ / Cmd+-** 调整字号
7. 关闭/移动窗口后下次启动自动恢复

## 开发

详见 [CLAUDE.md](CLAUDE.md)，包含：

- 常用命令（`make test` / `make build` / `make test-staging`）
- 开发 SOP（每次代码修改后必须立即构建并启动测试）
- 开发 → 测试 → 发布完整流程
- 关键约束（行号渲染、空行显示、跨行选择等踩过的坑）
- Wails v2.12.0 已知 API 差异

快速开始：

```bash
make test                  # 跑 Go 单元测试
PATH="$HOME/go/bin:$PATH" make test-staging   # 编译并启动到 /tmp/yiwo-test/
PATH="$HOME/go/bin:$PATH" make build          # 编译到 build/bin/
make package              # 出 .app + .zip
```

## 技术栈

- **后端** Go 1.22+ / Wails v2 / go-git / diff-match-patch / fsnotify / yaml.v3
- **前端** 原生 ES Modules + CSS Grid（无框架、无构建步骤）
- **架构** Wails 把 Go 后端和 WebKit 前端装在同一个进程，通过类型化 IPC（`App.*()`）通信。后端负责文件读取、Git 操作、diff 计算、YAML 解析、字数统计；前端负责渲染 UI 和捕获事件。

## 项目结构

```
main.go                          Wails 入口 + macOS 菜单 + 版本注入
setting.yaml                     默认配置
Makefile                         dev/test/build/package + ldflags
go.mod, go.sum
internal/
  model/types.go                 共享类型（DiffSegment, Commit, Result）
  config/load.go                 三段式配置（含 WindowState 持久化）
  file/                          加载 + Frontmatter + 字数（注释剔除）
  diff/                          diff-match-patch 包装 + 注释切分
  git/                           go-git 集成
  app/service.go                 Wails service + fsnotify + 后端 diff 编排
frontend/
  index.html
  css/                           theme / layout / diff / list / wordcount / toolbar / toast
  js/
    main.js, store.js, api.js
    views/  × 5                  mainView / commitList / wordcount / toolbar / toast
    util/  × 3                   clipboard / html / shortcuts
docs/superpowers/
  specs/                         设计规范
  plans/                         实施计划
```

## 路线图

v0.1 已发布（v1.0.0 – v1.2.0）。后续可能方向：

- Windows / Linux 平台支持
- 暗色模式
- 在应用内编辑 + 提交修改
- 多文件 Tab 切换
- 自定义徽章颜色 / 字体大小 UI

## 贡献

欢迎 PR。提交前请：

1. 跑 `make test` 确保所有 Go 测试通过
2. 跑 `PATH="$HOME/go/bin:$PATH" make test-staging` 在 `/tmp/yiwo-test/` 验证 UI 行为
3. 遵循 [CLAUDE.md](CLAUDE.md) 的关键约束（特别是行号渲染、空行显示、跨行选择相关）

## 许可证

[MIT License](LICENSE)

## 致谢

- [Wails](https://wails.io/) — Go + Web 桌面框架
- [diff-match-patch](https://github.com/sergi/go-diff) — Google 的字符级 diff 库
- [go-git](https://github.com/go-git/go-git) — 纯 Go Git 实现
