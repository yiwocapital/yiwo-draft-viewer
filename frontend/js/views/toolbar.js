import { getState, setState, subscribe } from "../store.js";
import { api } from "../api.js";
import { showToast } from "./toast.js";
import { copyText } from "../util/clipboard.js";
import { fixPunctuation } from "../util/punctuation.js";

export function init() {
  const toolbar = document.getElementById("toolbar");
  const foldCheckbox = document.getElementById("fold-comments-checkbox");
  const hideDiffCheckbox = document.getElementById("hide-diff-checkbox");

  function refresh() {
    const s = getState();
    toolbar.querySelectorAll("button").forEach((btn) => {
      const sec = btn.dataset.section;
      if (sec === "title" || sec === "summary") {
        btn.disabled = !s.hasFrontmatter;
      } else {
        btn.disabled = !s.loaded;
      }
    });
  }

  if (foldCheckbox) {
    foldCheckbox.addEventListener("change", async () => {
      const enabled = foldCheckbox.checked;
      setState({ foldComments: enabled });
      await api.setFoldComments(enabled);
      // Re-fetch diff so the backend's stripComments takes effect
      if (window.__loadDiff) window.__loadDiff();
    });
    subscribe((s) => {
      if (foldCheckbox.checked !== s.foldComments) {
        foldCheckbox.checked = s.foldComments;
      }
    });
  }

  if (hideDiffCheckbox) {
    hideDiffCheckbox.addEventListener("change", () => {
      setState({ hideDiff: hideDiffCheckbox.checked });
      // Re-fetch so backend returns static+content (or diff segments)
      if (window.__loadDiff) window.__loadDiff();
    });
    subscribe((s) => {
      if (hideDiffCheckbox.checked !== s.hideDiff) {
        hideDiffCheckbox.checked = s.hideDiff;
      }
    });
  }

  const editBtn = document.getElementById("edit-toggle-btn");
  if (editBtn) {
    editBtn.addEventListener("click", async () => {
      const s = getState();
      if (s.editMode) {
        window.dispatchEvent(new CustomEvent("yiwo-exit-edit"));
      } else {
        window.dispatchEvent(new CustomEvent("yiwo-enter-edit"));
      }
    });

    subscribe((s) => {
      if (s.editMode) {
        editBtn.textContent = s.dirty ? "● 编辑" : "✓ 编辑";
        editBtn.classList.add("active");
      } else {
        editBtn.textContent = "编辑";
        editBtn.classList.remove("active");
      }
      editBtn.disabled = !s.loaded;
    });
  }

  toolbar.addEventListener("click", async (e) => {
    const btn = e.target.closest("button");
    if (!btn || btn.disabled) return;
    if (!btn.dataset.section) return;
    const res = await api.copySection(btn.dataset.section);
    if (res.ok) {
      const fixed = fixPunctuation(res.data.text);
      await copyText(fixed);
      showToast("已复制");
    } else {
      showToast(`复制失败：${res.error}`, { kind: "error" });
    }
  });

  subscribe(refresh);
}
