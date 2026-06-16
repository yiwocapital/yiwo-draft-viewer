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

// Drag & drop
const app = document.getElementById("app");
app.addEventListener("dragover", (e) => { e.preventDefault(); });
app.addEventListener("drop", async (e) => {
  e.preventDefault();
  const f = e.dataTransfer.files[0];
  if (!f) return;
  if (f.name.toLowerCase().endsWith(".md") === false) {
    showToast("请拖入 .md 文件");
    return;
  }
  await openFile(f.path);
});

async function openFile(path) {
  const res = await api.openFile(path);
  if (!res.ok) {
    showToast(res.error);
    return;
  }
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
    if (cRes.data.items.length > 0) {
      window.__loadDiff && window.__loadDiff();
    }
  }
}

window.__openFile = openFile;
