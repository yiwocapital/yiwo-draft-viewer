import { getState, subscribe } from "../store.js";
import { escapeHtml } from "../util/html.js";

function renderSegments(segments) {
  return segments
    .map((s) => {
      const safe = escapeHtml(s.text);
      if (s.op === 1) return `<span class="diff-ins">${safe}</span>`;
      if (s.op === 2) return `<span class="diff-del">${safe}</span>`;
      return safe;
    })
    .join("");
}

function currentHeader(s) {
  if (s.selected === "WORKING") {
    return `<div class="version-header"><div class="subject">未提交</div><div class="body">工作区相对最新 commit 的差异</div></div>`;
  }
  const c = s.commits.find((x) => x.hash === s.selected);
  if (!c) return "";
  const lines = c.message.split("\n");
  const subject = lines[0] || "";
  const body = lines.slice(1).join("\n").trim();
  return `<div class="version-header">
    <div class="subject">${escapeHtml(subject)}</div>
    ${body ? `<div class="body">${escapeHtml(body)}</div>` : ""}
  </div>`;
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
    if (!s.loaded) {
      view.innerHTML = `<div class="empty-hint">拖一个 .md 文件进来，或菜单 File → Open</div>`;
      return;
    }
    if (s.diff.static) {
      view.innerHTML = `<pre class="static">${escapeHtml(s.content)}</pre>`;
      return;
    }
    if (!s.diff.segments || s.diff.segments.length === 0) {
      view.innerHTML = `<pre class="static">${escapeHtml(s.content)}</pre>`;
      return;
    }
    const header = currentHeader(s);
    view.innerHTML = `${header}<pre class="diff">${renderSegments(s.diff.segments)}</pre>`;
    scrollToFirstDiff(view);
  }

  subscribe(refresh);
}
