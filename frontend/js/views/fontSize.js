import { getState, subscribe } from "../store.js";

export function init() {
  const view = document.getElementById("main-view");

  function apply(s) {
    view.style.setProperty("--reading-font-size", `${s.fontSize}px`);
    // Force mainView re-render so the visual change is immediate
    if (window.__mainViewRefresh) window.__mainViewRefresh();
  }

  subscribe(apply);
  apply(getState());
}
