import { getState, setState } from "../store.js";
import { api } from "../api.js";
import { copyText } from "./clipboard.js";
import { showToast } from "../views/toast.js";

function applyFontSize(size) {
  // Set on documentElement so ALL descendants inherit
  document.documentElement.style.setProperty("--reading-font-size", `${size}px`);
  // Force mainView to re-render so line-numbered rows rebuild at new size
  if (window.__mainViewRefresh) window.__mainViewRefresh();
}

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
      applyFontSize(next);                    // immediate DOM update
      api.setFontSize(next);                  // fire-and-forget persistence
      setState({ fontSize: next });
    } else if ((e.metaKey || e.ctrlKey) && e.key === "-") {
      e.preventDefault();
      const cur = getState().fontSize;
      const next = Math.max(10, cur - 1);
      applyFontSize(next);
      api.setFontSize(next);
      setState({ fontSize: next });
    } else if ((e.metaKey || e.ctrlKey) && e.key === "f") {
      // Cmd+F: open search bar (or focus input if already open).
      // The input field handles Enter / Shift+Enter / Esc on its own.
      e.preventDefault();
      if (window.__search) window.__search.open();
    } else if ((e.metaKey || e.ctrlKey) && e.key === "g" && !e.shiftKey) {
      // Cmd+G: next match. Only meaningful when search is open with a term.
      if (window.__search && s.search.open && s.search.term) {
        e.preventDefault();
        window.__search.next();
      }
    } else if ((e.metaKey || e.ctrlKey) && e.key === "g" && e.shiftKey) {
      // Shift+Cmd+G: previous match.
      if (window.__search && s.search.open && s.search.term) {
        e.preventDefault();
        window.__search.prev();
      }
    } else if (e.key === "Escape" && !e.metaKey && !e.ctrlKey) {
      // Esc: close search if open. Plain Esc only — modifier-combos are
      // typically consumed by the OS / menu bar.
      if (window.__search && s.search.open) {
        e.preventDefault();
        window.__search.close();
      }
    }
  });
}
