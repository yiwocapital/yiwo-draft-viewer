import { getState, setState, subscribe } from "../store.js";

// Search bar controller. Owns:
//   - the search bar DOM (already in index.html, hidden by default)
//   - input → term mapping into store
//   - prev / next / close button behavior
//   - keyboard (Enter / Shift+Enter / Esc inside the input)
//   - scroll-to-current-match decision (skipped while user is typing)
//
// mainView.js owns the highlight DOM mutation — this file only decides WHEN
// to scroll and updates its own UI. Counters / disabled states come from
// reading .search-match spans in the rendered DOM.

let lastTerm = "";

export function init() {
  const bar = document.getElementById("search-bar");
  const input = document.getElementById("search-input");
  const counter = document.getElementById("search-counter");
  const prevBtn = document.getElementById("search-prev");
  const nextBtn = document.getElementById("search-next");
  const closeBtn = document.getElementById("search-close");

  function navigate(delta) {
    const s = getState();
    const matches = document.querySelectorAll(".search-match");
    const count = matches.length;
    if (count === 0) return;
    const next = ((s.search.currentIndex + delta) % count + count) % count;
    setState({ search: { ...s.search, currentIndex: next } });
  }

  function refresh() {
    const s = getState();

    // Detect typing: input has focus AND term changed since last refresh.
    // The setState from input events fires synchronously before refresh, so
    // this flag is true only for the keystroke that just landed.
    const inputFocused = document.activeElement === input;
    const isTyping = inputFocused && s.search.term !== lastTerm;
    lastTerm = s.search.term;

    // Bar visibility
    bar.classList.toggle("hidden", !s.search.open);

    // Keep input.value in sync with store only when input isn't being typed
    // into (avoids clobbering the cursor while user is typing).
    if (!inputFocused && input.value !== s.search.term) {
      input.value = s.search.term;
    }

    // Read live match count from rendered DOM. mainView.refresh runs before
    // ours (subs order: main → search), so highlights are already applied.
    const matches = document.querySelectorAll(".search-match");
    const count = matches.length;

    // Counter + disabled buttons
    const showTerm = s.search.open && s.search.term.length > 0;
    if (showTerm && count > 0) {
      counter.textContent = `${s.search.currentIndex + 1} / ${count}`;
      counter.classList.remove("zero");
    } else {
      counter.textContent = "0 / 0";
      counter.classList.add("zero");
    }
    prevBtn.disabled = !showTerm || count === 0;
    nextBtn.disabled = !showTerm || count === 0;

    // Scroll-to-current decision. Skip while typing (avoid screen jumping
    // on every keystroke). Also skip when there's nothing to scroll to.
    if (!isTyping && showTerm && count > 0) {
      const cur = document.querySelector(".search-match-current");
      if (cur) cur.scrollIntoView({ behavior: "smooth", block: "center" });
    }
  }

  // Input → store. Reset currentIndex so the new term always starts at the
  // first match (predictable behavior; doesn't force a scroll — see refresh).
  input.addEventListener("input", (e) => {
    const s = getState();
    setState({ search: { ...s.search, term: e.target.value, currentIndex: 0 } });
  });

  // Enter / Shift+Enter inside the input: navigate. Esc inside input: close.
  input.addEventListener("keydown", (e) => {
    if (e.key === "Enter") {
      e.preventDefault();
      navigate(e.shiftKey ? -1 : 1);
    } else if (e.key === "Escape") {
      e.preventDefault();
      close();
    }
  });

  prevBtn.addEventListener("click", () => navigate(-1));
  nextBtn.addEventListener("click", () => navigate(1));
  closeBtn.addEventListener("click", () => close());

  function open() {
    const s = getState();
    if (s.search.open) {
      input.focus();
      input.select();
      return;
    }
    setState({ search: { ...s.search, open: true } });
    // Focus after the bar becomes visible (next frame so the input is
    // actually focusable — classList.toggle is synchronous but focus() on
    // a freshly-revealed element can race with the layout pass).
    requestAnimationFrame(() => {
      input.focus();
      input.select();
    });
  }

  function close() {
    setState({ search: { open: false, term: "", currentIndex: 0 } });
    input.value = "";
    // Return focus to the main view so global shortcuts (Cmd+C, arrows,
    // Cmd+G) behave as if user is in the document body.
    document.getElementById("main-view")?.focus?.();
  }

  // Expose for shortcuts.js — keeps the keyboard wiring in one place.
  window.__search = { open, close, next: () => navigate(1), prev: () => navigate(-1) };

  subscribe(refresh);
  refresh();
}
