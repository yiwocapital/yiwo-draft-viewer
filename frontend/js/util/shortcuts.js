import { getState, setState } from "../store.js";
import { api } from "../api.js";
import { copyText } from "./clipboard.js";
import { showToast } from "../views/toast.js";

export function initShortcuts() {
  document.addEventListener("keydown", async (e) => {
    const s = getState();
    if (e.key === "ArrowDown" || e.key === "ArrowUp") {
      if (s.commits.length === 0) return;
      e.preventDefault();
      const idx = s.commits.findIndex((c) => c.hash === s.selected);
      let next = idx < 0 ? 0 : idx + (e.key === "ArrowDown" ? 1 : -1);
      next = Math.max(0, Math.min(s.commits.length - 1, next));
      setState({ selected: s.commits[next].hash, multiSelect: [] });
    } else if ((e.metaKey || e.ctrlKey) && e.key === "c") {
      const sel = window.getSelection().toString();
      if (sel) {
        await copyText(sel);
        showToast("已复制");
      } else {
        const res = await api.copySection("all");
        if (res.ok) {
          await copyText(res.data.text);
          showToast("已复制");
        }
      }
    } else if ((e.metaKey || e.ctrlKey) && (e.key === "=" || e.key === "+")) {
      e.preventDefault();
      const cur = getState().fontSize;
      const next = Math.min(32, cur + 1);
      api.setFontSize(next).then((res) => {
        if (res.ok) setState({ fontSize: res.data.fontSize });
      });
    } else if ((e.metaKey || e.ctrlKey) && e.key === "-") {
      e.preventDefault();
      const cur = getState().fontSize;
      const next = Math.max(10, cur - 1);
      api.setFontSize(next).then((res) => {
        if (res.ok) setState({ fontSize: res.data.fontSize });
      });
    }
  });
}
