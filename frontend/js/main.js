import { getState, setState, subscribe } from "./store.js";
import { api } from "./api.js";
import { showToast } from "./views/toast.js";
import { init as initMain } from "./views/mainView.js";
import { init as initList } from "./views/commitList.js";
import { init as initWc } from "./views/wordcount.js";
import { init as initTb } from "./views/toolbar.js";
import { initShortcuts } from "./util/shortcuts.js";

initMain();
initList();
initWc();
initTb();
initShortcuts();

// Drag-and-drop is handled by Wails via OnFileDrop in main.go.
// The "file-dropped" event listener (below) opens the file.

async function openFile(path) {
  const res = await api.openFile(path);
  if (!res.ok) {
    showToast(res.error);
    return;
  }
  window.__currentPath = path;
  setState({
    loaded: true,
    hasGit: res.data.hasGit,
    hasFrontmatter: res.data.hasFrontmatter,
    title: res.data.title,
    summary: res.data.summary,
    content: res.data.content,
    charCount: res.data.charCount,
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
  window.runtime.EventsOn("reload", () => {
    // Trigger a manual refresh by re-opening current path
    if (window.__currentPath) {
      openFile(window.__currentPath);
    }
  });
  window.runtime.EventsOn("close-file", () => {
    // Reset to empty state
    document.getElementById("app").classList.add("empty");
    setState({
      loaded: false, hasGit: false, hasFrontmatter: false,
      title: "", summary: "", content: "", charCount: 0,
      commits: [], selected: null, multiSelect: [],
      diff: { segments: [], charCount: 0, static: false },
    });
  });
}

window.__openFile = openFile;