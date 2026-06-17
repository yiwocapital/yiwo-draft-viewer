# YiwoDraftViewer

macOS Markdown 稿件预览工具。Go + Wails v2 + 原生前端（无框架）。

## 常用命令

```bash
make test                  # Go 单元测试
make build                 # 构建 .app（自动注入 git tag 版本号）
make package               # 出 .app + .zip
~/go/bin/wails dev   # 开发模式（wails CLI 不在默认 PATH）
```

`wails build -m` 的 `-m` 标志跳过 `go mod tidy`，避免删除尚未被 import 的直接依赖。

## 发布新版本

```bash
# 1. 确保所有改动已 commit（除 docs/ 下的计划文档外）
git status

# 2. 创建 annotated tag
git tag -a v1.x.y -m "YiwoDraftViewer v1.x.y — <一句话变更说明>"

# 3. 重新构建（Makefile 自动把 tag 名注入到窗口标题）
make build

# 4. 发布到用户级 Applications
rsync -a --delete build/bin/yiwo-draft-viewer.app/ ~/Applications/yiwo-draft-viewer.app/
```

发布后窗口标题会显示 `Yiwo Draft Viewer v1.x.y`。无 tag 时显示 `Yiwo Draft Viewer (dev/<short-sha>)`。

要替换系统级 `/Applications`（所有用户可见），需要 `sudo`；用户级 `~/Applications` 无需密码。

## 项目结构

```
main.go                          Wails 入口 + macOS 菜单 + 版本注入
setting.yaml                     默认配置
Makefile                         dev/test/build/package + ldflags
go.mod, go.sum
internal/
  model/types.go                 共享类型
  config/load.go                 三段式配置（含 WindowState 持久化）
  file/                          加载 + Frontmatter + 字数
  diff/                          diff-match-patch 包装
  git/                           go-git 集成
  app/service.go                 Wails service + fsnotify 监听
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

## 关键约束

- **构建后再测试**：`build/bin/yiwo-draft-viewer.app` 是临时构建产物。如果你修改了代码但没 `make build`，`open build/bin/...` 打开的可能是旧 binary。本地测试时优先 `open ~/Applications/yiwo-draft-viewer.app`，或修改代码后 `make clean build`。
- **依赖管理**：直接依赖没被 `import` 前会被 `go mod tidy` 删掉。如果新增了还没用上的依赖，用 `wails build -m` 而不是 `wails build`，或暂时不要跑 tidy。
- **行号渲染**：`frontend/css/diff.css` 用 CSS Grid + `::before` counter 实现。`.diff-row::before` 注入数字，`.diff-content` 包内容。**不要**用 `.diff-row > *` 选择器强制 column 2 —— text nodes 不被 `> *` 选中，会泄露到 column 1（行号列）。
- **空行显示**：`renderWithLineNumbers` 不要用 `if (parts[i].length > 0)` 守卫 —— 空字符串也要 push 进去；`.diff-content` 加 `min-height: 1.6em`。
- **跨行选择不复制行号**：`copy` 事件监听器过滤 `^[ \t]*\d+[ \t]+/gm`，作为兜底（`::before` 的 `user-select: none` 在某些 WebKit 场景不够）。
- **菜单 Quit 注册**：必须 `AppMenu.Append(menu.AppMenu())`，否则 Cmd+Q 触发 macOS 错误音。
- **窗口大小位置持久化**：`internal/app/service.go` 的 `Startup` 用 `~/Library/Application Support/YiwoDraftViewer/`，不要用 `"."`（启动目录不稳定）。前端 resize + mouseup 事件触发 `WindowChanged()`；`OnBeforeClose` 兜底。
- **配置文件位置**：`~/Library/Application Support/YiwoDraftViewer/setting.local.yaml`，不要写项目根。

## 手动测试

`docs/superpowers/MANUAL_TEST_GUIDE.md` 列了 7 个测试场景。用真实稿件 `~/Downloads/20260616-高盛评估霍尔木兹海峡重启情景/高盛评估霍尔木兹海峡重启情景-逐字稿.md` 验证（5 commit + 完整 YAML + 14KB 内容）。

## Wails v2.12.0 已知 API 差异

- `menu.HelpMenu()` 不存在 —— 用 macOS 系统默认 Help
- `OnFileDrop` 不是 `options.App` 字段，是 `runtime.OnFileDrop(ctx, cb)` 函数
- `go-git` v5.19.x 的 `Log()` 回调返回 `*object.Commit`（不是 `*gogit.Object`），`Hash()` 用 `plumbing.NewHash`