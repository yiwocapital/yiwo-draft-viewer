import { getState, subscribe } from "../store.js";
import { escapeHtml, stripComments } from "../util/html.js";
import { applyTerm } from "../util/highlight.js";

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
      currentLine.push({ op: seg.op, text: parts[i], isComment: seg.isComment });
    }
  }
  if (currentLine.length > 0) lines.push(currentLine);

  return lines.map((lineSegs) => {
    const inner = lineSegs.map((s) => {
      const safe = escapeHtml(s.text);
      if (s.isComment) {
        // Comment fragment: layer gray + italic on top of insert/delete coloring
        let cls = "diff-comment";
        if (s.op === 1) cls += " diff-comment-ins";
        else if (s.op === 2) cls += " diff-comment-del";
        return `<span class="${cls}">${safe}</span>`;
      }
      if (s.op === 1) return `<span class="diff-ins">${safe}</span>`;
      if (s.op === 2) return `<span class="diff-del">${safe}</span>`;
      return safe;
    }).join("");
    return `<div class="diff-row"><div class="diff-content">${inner}</div></div>`;
  }).join("");
}

function renderStaticWithLineNumbers(content) {
  const commentRE = /<!--[\s\S]*?-->/g;
  return content
    .split("\n")
    .map((line) => {
      // Highlight comments in gray; escape everything
      let html = "";
      let last = 0;
      for (const m of line.matchAll(commentRE)) {
        html += escapeHtml(line.slice(last, m.index));
        html += `<span class="diff-comment">${escapeHtml(line.slice(m.index, m.index + m[0].length))}</span>`;
        last = m.index + m[0].length;
      }
      html += escapeHtml(line.slice(last));
      return `<div class="diff-row"><div class="diff-content">${html}</div></div>`;
    })
    .join("");
}

// stripCommentsInLines removed — comment stripping now happens on the backend
// (in Go) before diff.Compute runs, so segments returned to the frontend are
// guaranteed to be free of comment fragments.

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
    // Status bar
    if (s.loaded && s.path) {
      if (statusBar.textContent !== s.path) statusBar.textContent = s.path;
      statusBar.classList.remove("hidden");
    } else {
      statusBar.textContent = "";
      statusBar.classList.add("hidden");
    }

    // Capture scroll position BEFORE innerHTML rewrite (innerHTML reset
    // drops scroll position to 0).
    const prevScrollTop = view.scrollTop;

    // Main content
    if (!s.loaded) {
      view.innerHTML = `<div class="empty-hint">拖一个 .md 文件进来，或菜单 File → Open</div>`;
      return;
    }

    // Two independent checkboxes determine what to render:
    //   hideDiff    — render selected commit's plain text instead of diff segments
    //   foldComments — strip <!-- --> from whatever we render
    // Four combinations, mapped below. The "selected commit's plain text" comes
    // from `s.diff.content` (populated by GetDiff when hideDiff=true); the
    // fallback (selected===null OR static+empty segments) uses `s.content` —
    // the working tree — since there's no specific commit to render.
    let body;
    const banner = multiSelectBanner(s);

    if (s.selected === null) {
      // No commit selected (file with no git history, or never opened): render
      // working tree content as plain text. Backend didn't strip — do it here.
      const content = s.foldComments ? stripComments(s.content) : s.content;
      body = renderStaticWithLineNumbers(content);
    } else if (s.hideDiff) {
      // "隐藏对比" on: show the selected commit's content (already stripped by
      // backend when foldComments was on at GetDiff time).
      const content = s.diff.content || "";
      body = renderStaticWithLineNumbers(content);
    } else if (s.diff.static || !s.diff.segments || s.diff.segments.length === 0) {
      // Empty diff fallback (e.g. the selected commit's blob is empty). Mirror
      // the no-commit-selected path: render working tree content stripped locally.
      const content = s.foldComments ? stripComments(s.content) : s.content;
      body = renderStaticWithLineNumbers(content);
    } else {
      // Normal diff mode — segments were already stripped by Go (because
      // foldComments is set on the backend before GetDiff was called).
      body = renderWithLineNumbers(s.diff.segments);
    }

    view.innerHTML = `${banner}<div class="diff">${body}</div>`;
    view.scrollTop = prevScrollTop;

    // Apply search highlights on top of the freshly-rendered diff/static
    // content. applyTerm operates on .diff-content blocks (which both
    // renderWithLineNumbers and renderStaticWithLineNumbers produce). For
    // empty term it just clears any prior highlights and returns count=0.
    applyTerm(view, s.search.term, s.search.currentIndex);

    // Only auto-scroll to first diff on the initial render. If the user
    // had already scrolled (prevScrollTop > 0), preserve their position.
    // Skip auto-scroll in static / hideDiff paths — there's no "first diff".
    if (prevScrollTop === 0 && !s.hideDiff && s.selected !== null &&
        !(s.diff.static || !s.diff.segments || s.diff.segments.length === 0)) {
      scrollToFirstDiff(view);
    }
  }

  subscribe(refresh);
  window.__mainViewRefresh = refresh;
}
