# 更新日志

所有重要变更都记录在此文件中。

格式参考 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)。

## [1.5.0] - 2026-06-23

### 新增
- **关键词搜索**：主视图按 `Cmd+F` 打开搜索栏，输入关键词即时高亮所有匹配、自动定位到当前匹配；`Cmd+G` / `Shift+Cmd+G` 在匹配间跳转；输入框内 `Enter` / `Shift+Enter` 也能跳转；`Esc` 关闭搜索
- **所见即所搜**：搜索匹配当前可见内容（含绿底新增 / 红底删除 / 灰色注释）；勾选「隐藏注释」或「隐藏对比」后，搜索自动按新可见内容重新匹配；切换 commit 时搜索词保留、自动跳到第一个匹配

## [1.4.0] - 2026-06-20

### 新增
- **「隐藏对比」开关**：右栏新增独立勾选框，开启后对所有 commit 都直接显示原文（不显示 diff 高亮），关闭时按 diff 模式渲染。和「隐藏注释」独立组合生效：两个开关同时开启 → 原文且无注释；只开「隐藏对比」→ 原文保留注释；只开「隐藏注释」→ 渲染 diff 但剥掉注释；都不开 → 正常 diff + 注释灰色显示

### 调整
- **「折叠编辑注释」改名为「隐藏注释」**：与新增的「隐藏对比」保持命名一致，且两个开关都做成胶囊样式（圆角、勾选时蓝底白字）并排一行，宽度对齐下方 4 个复制按钮
- **「通读全文」按钮和 `Esc` 快捷键移除**：原"强制把 selection 清空、显示 working tree 内容"的语义已拆为「隐藏对比」+「隐藏注释」两个全局开关。`Esc` 不再绑定任何操作

## [1.3.1] - 2026-06-18

### 修复
- 窗口标题规则强制执行：`make test-staging` 现在永远只显示 commit id（如 `Yiwo Draft Viewer (77e8d8f)`），不再误显示 tag。修复了 `git tag v1.x.y` 之后 `make test-staging` 因 HEAD 正好在 tag 上导致 `ReleaseVersion` 被错误注入、staging 窗口显示成发布版标题（`Yiwo Draft Viewer v1.x.y`）的问题。Makefile 拆出 `TEST_LDFLAGS`（不含 `ReleaseVersion`）专供 staging 使用，build target 继续用 `LDFLAGS`（含 `ReleaseVersion`），并在 CLAUDE.md 写入执行标准

## [1.3.0] - 2026-06-18

### 新增
- **HTML 编辑注释自动剔除**：字数统计不再包含 `<!-- -->` 编辑笔记中的字符
- **「折叠编辑注释」开关**：右栏复制按钮上方新增开关，开启后隐藏预览中的编辑注释，关闭时以浅灰色斜体显示
- **隔离测试构建**：新增 `make staging` 和 `make test-staging`，构建到 `/tmp/yiwo-test/` 临时位置并启动，与正式发布版完全隔离
- **开发 SOP 写入文档**：CLAUDE.md 中明确"每次代码修改后必须立即构建并启动测试"的强制流程
- **窗口标题自动显示版本**：已发布版本显示 `Yiwo Draft Viewer v<tag>`，未发布版本显示 `Yiwo Draft Viewer (<commit>)`
- **复制时英文标点自动转中文**：复制标题/摘要/正文/全部时，把紧邻中文的英文标点（`!`、`?`、`,`、`:`、`;`、`"`、`"`）转中文标点，仅修改剪贴板内容，原稿件不动
- **多选 commit 对比模式**：Cmd+点击 commit 切换为橙色高亮，单击显示与最新 commit 对比，再 Cmd+点击另一个锁定为固定对比
- **通读模式**：右栏「📖 通读全文」按钮或按 `Esc` 切换纯文本浏览视图
- **状态栏**：稿件预览下方显示当前文件的绝对路径，点击复制
- **Toast 通知分类**：复制成功 2.5 秒自动消失，错误需手动关闭
- **窗口大小/位置记忆**：下次启动自动恢复（事件驱动保存，不再轮询）
- **行号 + 段级 diff 高亮**：行级绿色新增/红色删除，同一行内多色混合
- **README + CHANGELOG**：项目根目录新增标准开源风格的说明文档

### 修复
- 跨 commit 对比：单选 vs latest 与双选固定对比的左右参数顺序不一致，导致颜色方向相反
- 注释的修改颜色：当注释内容被修改时，diff-match-patch 切碎注释块导致 stripComments 失败
- 切换「折叠编辑注释」时滚动位置跳到顶部
- 切换「折叠编辑注释」时不会立即刷新（loadDiff 缓存 key 没有包含 foldComments 状态）
- `Cmd+Q` 退出时触发 macOS 错误音（缺少 `menu.AppMenu()` 注册）
- 文件修改自动刷新失效（前端监听 `"reload"` 与后端发射的 `"reloaded"` 事件名不匹配）
- 窗口状态持久化路径改为 `~/Library/Application Support/YiwoDraftViewer/`，避免从不同目录启动时找不到配置
- 切 commit 时频繁闪烁（loadDiff 反复触发）
- 切换 commit 后多选状态残留 3 个高亮（多选逻辑未正确重置）

## [1.2.0] - 2026-06-18

### 新增
- **字数统计剔除编辑注释**：自动忽略 `<!-- -->` 编辑笔记中的字符
- **「折叠编辑注释」开关**：右栏复制按钮上方的新开关，开启后隐藏预览中的编辑注释
- **注释灰色显示**：关闭开关时，编辑注释以浅灰色斜体显示，与正文明显区分
- **隔离测试构建**：新增 `make staging` 和 `make test-staging`，编译到 `/tmp/yiwo-test/` 临时位置，不影响正式发布版
- **开发 SOP 写入文档**：CLAUDE.md 中明确了"每次代码修改后必须立即构建并启动测试"的强制流程

### 修复
- 跨 commit 对比：单选对比最新 vs 双选固定对比，现在两种模式的结果视觉一致
- 窗口状态持久化：配置路径改为 `~/Library/Application Support/YiwoDraftViewer/`，无论从哪个目录启动都能找到
- 窗口状态保存：从原来的 2 秒轮询改为事件驱动（拖动结束、关闭前自动保存）
- `Cmd+Q` 退出正常（之前因为 macOS 菜单未注册会触发错误音）
- 文件修改自动刷新预览（修复前后端事件名不匹配的问题）

## [1.1.0] - 2026-06-17

### 新增
- **窗口标题自动显示版本号**：发布版本显示 `Yiwo Draft Viewer v<tag>`，未发布版本显示 `Yiwo Draft Viewer (<commit>)`
- **Cmd+点击 commit 对比最新**：单击 commit 蓝色高亮显示与父 commit 的差异；Cmd+点击切换为橙色高亮，显示与最新 commit 的差异
- **多选对比**：再 Cmd+点击第二个 commit 锁定为两个 commit 之间的固定对比
- **复制时标点符号纠错**：复制标题/摘要/正文/全部时，自动把英文标点（如 `!`、`?`、`,`、`:`、`;`）转中文标点（`！`、`？`、`，`、`：`、`；`），仅修改剪贴板内容，原稿件不动
- **状态栏**：稿件预览下方显示当前文件的绝对路径，点击复制
- **「通读全文」模式**：右栏新按钮（或按 `Esc`），切换到不带 diff 高亮的纯文本浏览视图
- **字号快捷键**：`Cmd++` / `Cmd+-` 调整预览区字号，配置自动持久化
- **Toast 提示分类**：复制成功 2.5 秒自动消失，错误需手动关闭
- **窗口大小/位置记忆**：下次启动自动恢复
- **拖拽打开**：直接拖入 .md 文件即可打开
- **行级 + 段级 diff 高亮**：行号右侧显示绿色（新增）/ 红色删除线（删除），同一行内多色混合

### 修复
- 切换 commit 时不再"闪烁"
- 注释块（`<!-- -->`）现在不影响字数
- 多行文字在 diff 视图中左对齐，不再缩进

## [1.0.0] - 2026-06-17

首个公开预览版本。

### 核心功能
- 单 .md 文件预览与行号
- 字级 diff 高亮（绿色新增 / 红色删除）
- Git 历史浏览与多选对比
- 字数统计（折叠空白，rune 字符级）
- 结构化复制（标题 / 摘要 / 正文 / 全部）
- 文件修改自动刷新
- 拖拽或菜单打开
- 原生 macOS 菜单栏
- 通读模式

[1.5.0]: https://github.com/yiwocapital/yiwo-draft-viewer/compare/v1.4.0...v1.5.0
[1.4.0]: https://github.com/yiwocapital/yiwo-draft-viewer/compare/v1.3.1...v1.4.0
[1.3.1]: https://github.com/yiwocapital/yiwo-draft-viewer/compare/v1.3.0...v1.3.1
[1.3.0]: https://github.com/yiwocapital/yiwo-draft-viewer/compare/v1.2.0...v1.3.0
[1.2.0]: https://github.com/yiwocapital/yiwo-draft-viewer/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/yiwocapital/yiwo-draft-viewer/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/yiwocapital/yiwo-draft-viewer/releases/tag/v1.0.0
