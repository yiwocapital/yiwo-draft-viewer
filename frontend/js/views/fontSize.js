import { getState, subscribe } from "../store.js";

function applyFontSize(size) {
  document.documentElement.style.setProperty("--reading-font-size", `${size}px`);
}

export function init() {
  // Apply initial value from store
  applyFontSize(getState().fontSize);
  // Keep in sync with future changes
  subscribe((s) => applyFontSize(s.fontSize));
}
