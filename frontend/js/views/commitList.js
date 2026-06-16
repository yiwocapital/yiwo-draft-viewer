import { getState, setState, subscribe } from "../store.js";
import { api } from "../api.js";
import { copyText } from "../util/clipboard.js";
import { showToast } from "./toast.js";

function timeAgo(ts) {
  const diff = Math.floor(Date.now() / 1000 - ts);
  if (diff < 60) return "刚刚";
  if (diff < 3600) return `${Math.floor(diff / 60)}分钟前`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}小时前`;
  return `${Math.floor(diff / 86400)}天前`;
}

function escapeHtml(s) {
  return s.replace(/[&<>"']/g, (m) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[m]));
}

function renderItem(c, selected) {
  const div = document.createElement("div");
  div.className = "commit-item";
  if (c.isUnstaged) div.classList.add("unstaged");
  if (selected) div.classList.add("selected");
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

export function init() {
  const list = document.getElementById("commit-list");

  function refresh() {
    const s = getState();
    list.innerHTML = "";
    if (!s.hasGit) {
      const empty = document.createElement("div");
      empty.className = "commit-list-empty";
      empty.textContent = "未启用版本控制";
      list.appendChild(empty);
      return;
    }
    s.commits.forEach((c) => {
      const isSel = s.selected === c.hash;
      const item = renderItem(c, isSel);
      item.addEventListener("click", (e) => {
        if (e.target.classList.contains("copy-hash")) {
          copyText(c.hash).then(() => showToast("已复制"));
          return;
        }
        if (e.metaKey || e.ctrlKey) {
          const multi = [...s.multiSelect];
          const idx = multi.indexOf(c.hash);
          if (idx >= 0) multi.splice(idx, 1);
          else if (multi.length < 2) multi.push(c.hash);
          else multi[0] = c.hash;
          setState({ multiSelect: multi });
        } else {
          setState({ selected: c.hash, multiSelect: [] });
          loadDiff();
        }
      });
      list.appendChild(item);
    });
  }

  async function loadDiff() {
    const s = getState();
    if (!s.selected) return;
    if (s.multiSelect.length === 2) {
      const [a, b] = s.multiSelect;
      const res = await api.getDiff(a, b);
      if (res.ok) setState({ diff: res.data });
    } else {
      const idx = s.commits.findIndex((c) => c.hash === s.selected);
      if (idx < 0) return;
      const cur = s.commits[idx].hash;
      const prev = idx + 1 < s.commits.length ? s.commits[idx + 1].hash : "";
      const res = await api.getDiff(prev, cur);
      if (res.ok) setState({ diff: res.data });
    }
  }

  subscribe(refresh);
  subscribe(loadDiff);
  window.__loadDiff = loadDiff; // for keyboard nav
}
