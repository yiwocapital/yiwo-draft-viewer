import { getState, subscribe } from "../store.js";

export function init() {
  const view = document.getElementById("main-view");
  subscribe((s) => {
    view.style.setProperty("--reading-font-size", `${s.fontSize}px`);
  });
}
