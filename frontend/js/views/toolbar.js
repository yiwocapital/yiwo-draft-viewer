import { getState, setState, subscribe } from "../store.js";
import { api } from "../api.js";
import { showToast } from "./toast.js";
import { copyText } from "../util/clipboard.js";
import { fixPunctuation } from "../util/punctuation.js";

export function init() {
  const toolbar = document.getElementById("toolbar");
  const foldCheckbox = document.getElementById("fold-comments-checkbox");

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
    foldCheckbox.addEventListener("change", () => {
      setState({ foldComments: foldCheckbox.checked });
    });
    subscribe((s) => {
      if (foldCheckbox.checked !== s.foldComments) {
        foldCheckbox.checked = s.foldComments;
      }
    });
  }

  toolbar.addEventListener("click", async (e) => {
    const btn = e.target.closest("button");
    if (!btn || btn.disabled) return;
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
