import { getState, setState, subscribe } from "../store.js";
import { api } from "../api.js";
import { copyText } from "../util/clipboard.js";
import { showToast } from "./toast.js";
import { escapeHtml } from "../util/html.js";

function timeAgo(ts) {
  const diff = Math.floor(Date.now() / 1000 - ts);
  if (diff < 60) return "刚刚";
  if (diff < 3600) return `${Math.floor(diff / 60)}分钟前`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}小时前`;
  return `${Math.floor(diff / 86400)}天前`;
}

function renderItem(c, selected, multiSelected) {
  const div = document.createElement("div");
  div.className = "commit-item";
  if (c.isUnstaged) div.classList.add("unstaged");
  if (selected) div.classList.add("selected");
  if (multiSelected) div.classList.add("multi-selected");
  div.dataset.hash = c.hash;

  const firstLine = c.firstLine + (c.hasMore ? " ..." : "");
  div.innerHTML = `
    <span class="commit-hash">${escapeHtml(c.shortHash)}</span>
    <span class="copy-hash" title="复制完整哈希">📋</span>
    <div class="commit-msg">${escapeHtml(firstLine)}</div>
    <div class="commit-time">${timeAgo(c.timestamp)}</div>
  `;
  return div;
}

function renderDetail(s) {
  const detail = document.getElementById("commit-detail");
  if (!detail) return;
  if (!s.hasGit || !s.selected || s.commits.length === 0) {
    detail.innerHTML = "";
    detail.classList.add("hidden");
    return;
  }
  const c = s.commits.find((x) => x.hash === s.selected);
  if (!c) {
    detail.innerHTML = "";
    detail.classList.add("hidden");
    return;
  }
  detail.classList.remove("hidden");
  const subject = c.message.split("\n")[0] || "";
  const body = c.message.split("\n").slice(1).join("\n").trim();
  detail.innerHTML = `
    <div class="detail-subject">${escapeHtml(subject)}</div>
    ${body ? `<div class="detail-body">${escapeHtml(body)}</div>` : ""}
    <div class="detail-meta">
      <span class="detail-hash" title="点击复制">${escapeHtml(c.hash)}</span>
      <span class="detail-time">${new Date(c.timestamp * 1000).toLocaleString("zh-CN")}</span>
    </div>
  `;
}

let lastLoadKey = null;

async function loadDiff() {
  const s = getState();
  if (!s.selected) return;

  let key;
  let args;
  if (s.multiSelect.length === 2) {
    const [a, b] = s.multiSelect;
    key = `multi:${a}:${b}`;
    args = [a, b];
  } else {
    const idx = s.commits.findIndex((c) => c.hash === s.selected);
    if (idx < 0) return;
    const cur = s.commits[idx].hash;
    const prev = idx + 1 < s.commits.length ? s.commits[idx + 1].hash : "";
    key = `single:${prev}:${cur}`;
    args = [prev, cur];
  }

  if (key === lastLoadKey) return;
  lastLoadKey = key;

  const res = await api.getDiff(args[0], args[1]);
  if (res.ok) setState({ diff: res.data, charCount: res.data.charCount });
}

export function init() {
  const list = document.getElementById("commit-list");
  const readMode = document.getElementById("read-mode");

  function refresh() {
    const s = getState();
    list.innerHTML = "";
    if (!s.hasGit) {
      const empty = document.createElement("div");
      empty.className = "commit-list-empty";
      empty.textContent = "未启用版本控制";
      list.appendChild(empty);
      renderDetail(s);
      if (readMode) readMode.classList.remove("active");
      return;
    }
    // Read-mode active state
    if (readMode) {
      if (s.selected === null) {
        readMode.classList.add("active");
      } else {
        readMode.classList.remove("active");
      }
    }
    s.commits.forEach((c) => {
      const isSel = s.selected === c.hash;
      const isMulti = s.multiSelect.includes(c.hash);
      list.appendChild(renderItem(c, isSel, isMulti));
    });
    renderDetail(s);
  }

  list.addEventListener("click", async (e) => {
    const item = e.target.closest(".commit-item");
    if (!item) return;
    const hash = item.dataset.hash;
    const s = getState();
    const c = s.commits.find((x) => x.hash === hash);
    if (!c) return;

    if (e.target.classList.contains("copy-hash")) {
      await copyText(c.hash);
      showToast("已复制");
      return;
    }
    if (e.metaKey || e.ctrlKey) {
      const multi = [...s.multiSelect];
      const idx = multi.indexOf(c.hash);
      if (idx >= 0) {
        // Clicked an already-multi-selected commit — remove it
        multi.splice(idx, 1);
      } else {
        // Add new commit; cap at 2 with FIFO (drop oldest)
        multi.push(c.hash);
        if (multi.length > 2) {
          multi.shift();
        }
      }
      // Update BOTH selected (for the highlight) and multiSelect
      setState({ selected: c.hash, multiSelect: multi });
      return;
    }
    if (s.selected === c.hash) return;
    setState({ selected: c.hash, multiSelect: [] });
  });

  if (readMode) {
    readMode.addEventListener("click", () => {
      setState({ selected: null, multiSelect: [] });
      lastLoadKey = null;
    });
  }

  subscribe(refresh);
  subscribe(loadDiff);
  window.__loadDiff = loadDiff;
}
