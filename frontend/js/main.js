import { getState, setState, subscribe } from "./store.js";
import { api } from "./api.js";
import { showToast } from "./views/toast.js";
import { init as initMain } from "./views/mainView.js";
import { init as initList } from "./views/commitList.js";
import { init as initWc } from "./views/wordcount.js";
import { init as initTb } from "./views/toolbar.js";
import { init as initFs } from "./views/fontSize.js";
import { init as initSearch } from "./views/search.js";
import { initShortcuts } from "./util/shortcuts.js";

initMain();
initList();
initWc();
initTb();
initFs();
initSearch();
initShortcuts();

// Drag-and-drop is handled by Wails via OnFileDrop in main.go.
// The "file-dropped" event listener (below) opens the file.

async function openFile(path) {
  const res = await api.openFile(path);
  if (!res.ok) {
    showToast(res.error, { kind: "error" });
    return;
  }
  window.__currentPath = path;
  setState({
    loaded: true,
    path: res.data.path,
    hasGit: res.data.hasGit,
    hasFrontmatter: res.data.hasFrontmatter,
    title: res.data.title,
    summary: res.data.summary,
    content: res.data.content,
    charCount: res.data.charCount,
    fontSize: res.data.fontSize || 14,
  });
  const cRes = await api.listCommits();
  if (cRes.ok) {
    setState({
      commits: cRes.data.items,
      selected: cRes.data.items[0]?.hash || null,
      multiSelect: [],
    });
  }
}

// Wails event listeners (from native macOS menu + drag-and-drop)
if (window.runtime && window.runtime.EventsOn) {
  window.runtime.EventsOn("open-file", (path) => {
    openFile(path);
  });
  window.runtime.EventsOn("file-dropped", (path) => {
    openFile(path);
  });
  // Reload logic shared by fsnotify "reloaded" event and menu Cmd+R "reload" event
  function reloadFromBackend() {
    if (!window.__currentPath) return;
    api.listCommits().then((cRes) => {
      if (!cRes.ok) return;
      const cur = getState().selected;
      const stillThere = cRes.data.items.find((c) => c.hash === cur);
      const nextSel = stillThere ? cur : (cRes.data.items[0]?.hash || null);
      setState({
        commits: cRes.data.items,
        selected: nextSel,
      });
    });
  }
  window.runtime.EventsOn("reloaded", reloadFromBackend);
  window.runtime.EventsOn("reload", reloadFromBackend);
  // OnBeforeClose intercepts Cmd+Q with unsaved edits and emits this event.
  // We previously used Wails runtime.MessageDialog here, but it silently
  // failed on some Wails v2 + macOS combos. Native JS confirm() works
  // reliably inside the webview, so the dirty-quit prompt now lives here.
  window.runtime.EventsOn("request-dirty-quit", async () => {
    const s = getState();
    // Race: dirty may have been cleared between OnBeforeClose firing and
    // this listener running (e.g., SaveAndClose completed in another path).
    if (!s.dirty) {
      api.quit();
      return;
    }

    // 3-option dialog using sequential confirm() calls (same pattern as
    // yiwo-conflict-detected). Native JS confirm() works reliably on macOS
    // webview; Wails MessageDialog from OnBeforeClose goroutine does not.
    const choice = await new Promise((resolve) => {
      // Step 1: 保存 or 取消 (to see next option)?
      const save = confirm("当前编辑有未保存修改。\n\n点击「确定」保存后退出。\n点击「取消」查看下一步选项。");
      if (save) { resolve("save"); return; }
      // Step 2: 丢弃 or 取消 (cancel and stay in app)?
      const discard = confirm("点击「确定」放弃修改并退出。\n点击「取消」留在 app。");
      if (discard) { resolve("discard"); return; }
      resolve("cancel");
    });

    if (choice === "save") {
      // SaveAndClose writes content via Service.Save then runtime.Quit.
      // If Save fails (e.g., EXTERNAL_MODIFIED), frontend should NOT quit —
      // user needs to resolve the conflict.
      await api.saveAndClose(s.content);
    } else if (choice === "discard") {
      // Clear backend dirty mirror so OnBeforeClose won't re-trigger
      // dirty prompt on the next close attempt.
      await api.setDirty(false);
      // Tell backend to quit.
      api.quit();
    }
    // "cancel": do nothing. App stays open. Backend's OnBeforeClose
    // already returned prevent=true, so the original close attempt was
    // intercepted. Next Cmd+Q will re-trigger this whole flow.
  });
  window.runtime.EventsOn("close-file", async () => {
    if (getState().dirty) {
      const ok = confirm("当前编辑有未保存修改，确定关闭？\n（修改将被丢弃）");
      if (!ok) return;
    }
    // Tell backend to clear state (this was previously done by the Go
    // menu callback; now it's frontend-driven so the dirty prompt can
    // run first).
    try {
      await api.closeFile();
    } catch (e) {
      // ignore — UI state still gets reset below
    }
    document.getElementById("app").classList.add("empty");
    setState({
      loaded: false, path: "", hasGit: false, hasFrontmatter: false,
      title: "", summary: "", content: "", charCount: 0,
      commits: [], selected: null, multiSelect: [],
      diff: { segments: [], charCount: 0, static: false },
      search: { open: false, term: "", currentIndex: 0 },
      editMode: false,
      dirty: false,
    });
  });
}

window.__openFile = openFile;

// Status bar: click-to-copy path
const statusBar = document.getElementById("status-bar");
if (statusBar) {
  statusBar.addEventListener("click", async () => {
    const path = getState().path;
    if (path) {
      const { copyText } = await import("./util/clipboard.js");
      const ok = await copyText(path);
      if (ok) showToast("已复制路径");
    }
  });
  statusBar.title = "点击复制路径";
}

// Initialize font size from config
api.getFontSize().then((size) => {
  if (typeof size === "number") setState({ fontSize: size });
});

// Persist window state on user-driven changes (resize + window-drag mouseup).
// Debounced so a continuous drag/resize coalesces into one save.
let winSaveTimer = null;
function scheduleWindowSave() {
  if (winSaveTimer) clearTimeout(winSaveTimer);
  winSaveTimer = setTimeout(() => {
    if (api.windowChanged) api.windowChanged();
    winSaveTimer = null;
  }, 400);
}
window.addEventListener("resize", scheduleWindowSave);
// Window-drag end: webview doesn't fire resize when the native window is moved
// without resizing. mouseup captures the moment the drag ends.
window.addEventListener("mouseup", scheduleWindowSave);

// Strip leading line numbers from copied text.
// `user-select: none` + `pointer-events: none` on `.diff-row::before` already
// prevents most selections, but WebKit sometimes folds the pseudo-element
// text into the selection rect. This is a defensive filter.
document.addEventListener("copy", (e) => {
  const sel = window.getSelection();
  if (!sel.rangeCount) return;
  const text = sel.toString();
  // Match leading "NNN " (digits + whitespace) at the start of each line.
  const filtered = text.replace(/^[ \t]*\d+[ \t]+/gm, "");
  if (filtered !== text) {
    e.clipboardData.setData("text/plain", filtered);
    e.preventDefault();
  }
});

// Edit-mode event wiring (Task 8): toolbar / shortcuts dispatch these
// CustomEvents; main.js is the single point that talks to the Go backend.
window.addEventListener("yiwo-enter-edit", async () => {
  const s = getState();
  if (s.editMode) return;
  const res = await api.beginEdit();
  if (!res.ok) {
    showToast(res.error || "无法进入编辑模式", { kind: "error" });
    return;
  }
  setState({ editMode: true, dirty: false });
});

window.addEventListener("yiwo-exit-edit", async () => {
  const s = getState();
  if (!s.editMode) return;
  if (s.dirty) {
    const ok = confirm("当前编辑有未保存修改，确定丢弃？");
    if (!ok) return;
  }
  // Tell the backend we're done (resumes fsnotify watcher); ignore result
  // since user is leaving edit mode regardless.
  await api.endEdit();
  await api.setDirty(false);
  setState({ editMode: false, dirty: false });
});

window.addEventListener("yiwo-save-edit", async () => {
  const s = getState();
  if (!s.editMode || !s.dirty) return;
  const res = await api.save(s.content);
  if (res.ok) {
    setState({ dirty: false });
    showToast("已保存");
  } else if (res.code === "EXTERNAL_MODIFIED") {
    // Handled in Task 9 (conflict dialog)
    showToast("外部已修改，请选择处理方式", { kind: "warning" });
    window.dispatchEvent(new CustomEvent("yiwo-conflict-detected"));
  } else {
    showToast(`保存失败：${res.error || "未知错误"}`, { kind: "error" });
  }
});

// Task 9: Conflict dialog — three resolution paths when Save detects the file
// was modified externally during edit. Uses two sequential native confirm()
// dialogs (v1): overwrite / reload / cancel.
window.addEventListener("yiwo-conflict-detected", async () => {
  const choice = await new Promise((resolve) => {
    // First: overwrite external?
    const overwrite = confirm("外部文件已修改。\n\n点击「确定」覆盖外部修改（外部改动将丢失）。\n点击「取消」查看下一步选项。");
    if (overwrite) {
      resolve("overwrite");
      return;
    }
    // Second: reload from disk?
    const reload = confirm("点击「确定」放弃本地修改，重新读入外部版本。\n点击「取消」留在编辑模式（保持本地修改）。");
    if (reload) {
      resolve("reload");
    } else {
      resolve("cancel");
    }
  });

  const s = getState();
  if (choice === "overwrite") {
    const res = await api.saveOverwrite(s.content);
    if (res.ok) {
      setState({ dirty: false });
      showToast("已覆盖外部修改");
    } else {
      showToast(`保存失败：${res.error || "未知错误"}`, { kind: "error" });
    }
  } else if (choice === "reload") {
    // Exit edit mode (endEdit triggers OpenFile internally), then refresh
    // the frontend store from disk so commit list and content update.
    await api.endEdit();
    await api.setDirty(false);
    setState({ editMode: false, dirty: false });
    if (window.__currentPath && window.__openFile) {
      await window.__openFile(window.__currentPath);
    }
    showToast("已重新读入外部版本");
  }
  // "cancel": do nothing, stay in edit mode
});