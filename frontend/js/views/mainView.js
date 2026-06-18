import { getState, subscribe } from "../store.js";
import { escapeHtml, stripComments } from "../util/html.js";

function renderWithLineNumbers(segments) {
  // Group segments by line (split at \n). Each line becomes a row; each row
  // contains 0+ segments, each preserving its original op for inline coloring.
  const lines = [];
  let currentLine = [];
  for (const seg of segments) {
    const parts = seg.text.split("\n");
    for (let i = 0; i < parts.length; i++) {
      if (i > 0) {
        lines.push(currentLine);
        currentLine = [];
      }
      // Always push, even when parts[i] === "" — preserves blank lines
      currentLine.push({ op: seg.op, text: parts[i] });
    }
  }
  if (currentLine.length > 0) lines.push(currentLine);

  return lines.map((lineSegs) => {
    const inner = lineSegs.map((s) => {
      const safe = escapeHtml(s.text);
      if (s.op === 1) return `<span class="diff-ins">${safe}</span>`;
      if (s.op === 2) return `<span class="diff-del">${safe}</span>`;
      return safe;
    }).join("");
    return `<div class="diff-row"><div class="diff-content">${inner}</div></div>`;
  }).join("");
}

function renderStaticWithLineNumbers(content) {
  return content
    .split("\n")
    .map((line) => `<div class="diff-row"><div class="diff-content">${escapeHtml(line)}</div></div>`)
    .join("");
}

function multiSelectBanner(s) {
  if (s.multiSelect.length === 0) return "";

  if (s.multiSelect.length === 1) {
    const sel = s.commits.find((c) => c.hash === s.multiSelect[0]);
    if (!sel) return "";
    const latest = s.commits[0];
    if (latest && latest.hash === sel.hash) {
      return `<div class="multi-banner">对比：${escapeHtml(sel.shortHash)} 与上一版（已是最新）</div>`;
    }
    return `<div class="multi-banner">对比：最新 ↔ ${escapeHtml(sel.shortHash)}</div>`;
  }

  // length === 2
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

  function refresh() {
    const s = getState();
    const statusBar = document.getElementById("status-bar");
    // Status bar (lives outside #main-view in the DOM, so survives innerHTML rewrite)
    if (s.loaded && s.path) {
      if (statusBar.textContent !== s.path) statusBar.textContent = s.path;
      statusBar.classList.remove("hidden");
    } else {
      statusBar.textContent = "";
      statusBar.classList.add("hidden");
    }

    // Strip comments if toggle is on
    let segments = s.diff.segments;
    let staticContent = s.content;
    if (s.foldComments) {
      if (segments) {
        segments = segments.map((seg) => ({
          ...seg,
          text: stripComments(seg.text),
        })).filter((seg) => seg.text.length > 0);
      }
      staticContent = stripComments(s.content);
    }

    // Main content
    if (!s.loaded) {
      view.innerHTML = `<div class="empty-hint">拖一个 .md 文件进来，或菜单 File → Open</div>`;
      return;
    }
    if (s.selected === null) {
      // Read mode — plain content (comments stripped if toggle on)
      view.innerHTML = `<div class="diff">${renderStaticWithLineNumbers(staticContent)}</div>`;
      return;
    }
    if (s.diff.static || !segments || segments.length === 0) {
      view.innerHTML = `${multiSelectBanner(s)}<div class="diff">${renderStaticWithLineNumbers(staticContent)}</div>`;
      return;
    }
    view.innerHTML = `${multiSelectBanner(s)}<div class="diff">${renderWithLineNumbers(segments)}</div>`;
    scrollToFirstDiff(view);
  }

  subscribe(refresh);
  window.__mainViewRefresh = refresh;
}
