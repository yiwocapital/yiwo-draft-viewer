import { getState, subscribe } from "../store.js";
import { escapeHtml } from "../util/html.js";

function renderWithLineNumbers(segments) {
  const lines = [];
  let currentLine = "";
  let currentOp = 0;

  for (const seg of segments) {
    const parts = seg.text.split("\n");
    for (let i = 0; i < parts.length; i++) {
      if (i > 0) {
        lines.push({ text: currentLine, op: currentOp });
        currentLine = parts[i];
        currentOp = seg.op;
      } else {
        if (currentOp === 0) currentOp = seg.op;
        currentLine += parts[i];
      }
    }
  }
  if (currentLine.length > 0 || lines.length === 0) {
    lines.push({ text: currentLine, op: currentOp });
  }

  return lines.map((line, i) => {
    const num = String(i + 1).padStart(3, " ");
    const safe = escapeHtml(line.text);
    let cls = "";
    if (line.op === 1) cls = "diff-ins";
    else if (line.op === 2) cls = "diff-del";
    return `<div class="diff-line${cls ? " " + cls : ""}"><span class="line-num">${num}</span><span class="line-text">${safe}</span></div>`;
  }).join("");
}

function renderStaticWithLineNumbers(content) {
  const lines = content.split("\n");
  return lines.map((line, i) => {
    const num = String(i + 1).padStart(3, " ");
    return `<div class="diff-line"><span class="line-num">${num}</span><span class="line-text">${escapeHtml(line)}</span></div>`;
  }).join("");
}

function multiSelectBanner(s) {
  if (s.multiSelect.length !== 2) return "";
  const [a, b] = s.multiSelect;
  const ca = s.commits.find((c) => c.hash === a);
  const cb = s.commits.find((c) => c.hash === b);
  if (!ca || !cb) return "";
  return `<div class="multi-banner">对比：${escapeHtml(ca.shortHash)} ↔ ${escapeHtml(cb.shortHash)}</div>`;
}

function scrollToFirstDiff(view) {
  const first = view.querySelector(".diff-ins, .diff-del");
  if (first) {
    first.scrollIntoView({ behavior: "smooth", block: "center" });
  }
}

export function init() {
  const view = document.getElementById("main-view");
  const statusBar = document.getElementById("status-bar");

  function refresh() {
    const s = getState();
    // Status bar (sticky, persistent element)
    if (s.loaded && s.path) {
      statusBar.textContent = s.path;
      statusBar.classList.remove("hidden");
    } else {
      statusBar.textContent = "";
      statusBar.classList.add("hidden");
    }
    // Main content
    if (!s.loaded) {
      view.innerHTML = `<div class="empty-hint">拖一个 .md 文件进来，或菜单 File → Open</div>`;
      return;
    }
    if (s.diff.static || !s.diff.segments) {
      view.innerHTML = `${multiSelectBanner(s)}<div class="diff">${renderStaticWithLineNumbers(s.content)}</div>`;
      return;
    }
    view.innerHTML = `${multiSelectBanner(s)}<div class="diff">${renderWithLineNumbers(s.diff.segments)}</div>`;
    scrollToFirstDiff(view);
  }

  subscribe(refresh);
  window.__mainViewRefresh = refresh;
}
