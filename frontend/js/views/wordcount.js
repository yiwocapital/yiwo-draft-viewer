import { getState, subscribe } from "../store.js";

export function init() {
  const el = document.getElementById("wordcount");
  const val = document.getElementById("wordcount-value");

  subscribe((s) => {
    if (s.loaded) {
      el.classList.remove("hidden");
      val.textContent = s.charCount.toLocaleString();
    } else {
      el.classList.add("hidden");
    }
  });
}
