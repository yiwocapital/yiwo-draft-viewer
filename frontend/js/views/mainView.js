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

// stripCommentsInLines splits segments by '\n', strips HTML comments from
// each line's joined text, drops empty lines, and re-merges into single-op
// segments. This catches comment fragments that diff-match-patch split apart.
function stripCommentsInLines(segments) {
  const lines = [];
  let currentLine = [];

  for (const seg of segments) {
    const parts = seg.text.split("\n");
    for (let i = 0; i < parts.length; i++) {
      if (i > 0) {
        lines.push(currentLine);
        currentLine = [];
      }
      if (parts[i].length > 0) {
        currentLine.push({ op: seg.op, text: parts[i] });
      }
    }
  }
  if (currentLine.length > 0) lines.push(currentLine);

  const out = [];
  for (const lineSegs of lines) {
    const joined = lineSegs.map((s) => s.text).join("");
    const stripped = stripComments(joined);
    if (stripped === "") continue; // entire line was a comment
    // Pick the dominant op (first non-equal), or equal if all are equal
    let mainOp = 0;
    for (const s of lineSegs) {
      if (s.op !== 0) {
        mainOp = s.op;
        break;
      }
    }
    out.push({ op: mainOp, text: stripped });
  }
  return out;
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

    // Main content
    if (!s.loaded) {
      view.innerHTML = `<div class="empty-hint">拖一个 .md 文件进来，或菜单 File → Open</div>`;
      return;
    }

    // Strip static content if foldComments
    const staticContent = s.foldComments ? stripComments(s.content) : s.content;

    if (s.selected === null) {
      view.innerHTML = `<div class="diff">${renderStaticWithLineNumbers(staticContent)}</div>`;
      return;
    }

    // Prepare diff segments. If foldComments is on, strip at the LINE level so
    // fragments of an HTML comment that diff-match-patch split apart are
    // correctly removed.
    const rawSegments = s.diff.segments;
    const useFolded = s.foldComments && rawSegments && rawSegments.length > 0;
    const segments = useFolded ? stripCommentsInLines(rawSegments) : rawSegments;

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
