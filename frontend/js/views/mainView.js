import { getState, subscribe } from "../store.js";
import { escapeHtml } from "../util/html.js";

function renderWithLineNumbers(segments) {
  // Walk segments, splitting at \n boundaries. Each "line" gets the op of its starting segment.
  const lines = [];
  let currentLine = "";
  let currentOp = 0;

  for (const seg of segments) {
    const parts = seg.text.split("\n");
    for (let i = 0; i < parts.length; i++) {
      if (i > 0) {
        // \n boundary — flush current line
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
      view.innerHTML = `<div class="diff">${renderStaticWithLineNumbers(s.content)}</div>`;
      return;
    }
    const header = currentHeader(s);
    view.innerHTML = `${header}<div class="diff">${renderWithLineNumbers(s.diff.segments)}</div>`;
    scrollToFirstDiff(view);
  }

  subscribe(refresh);
}