import { getState, setState } from "../store.js";

export function init() {
  // no-op for now; toolbar/main.js import this module for its exports
}

export function render() {
  const s = getState();
  // Defensive: should never be called outside edit mode, but if it is,
  // bail to empty.
  if (!s.editMode) return "";

  const fontSize = s.fontSize || 14;
  // Note: textarea.value is set imperatively after innerHTML (innerHTML
  // can't carry the textarea value without escaping headaches).
  return `<textarea id="editor-textarea" class="editor-textarea"
            style="font-size: ${fontSize}px;"
            spellcheck="false"
            autocomplete="off"
            autocorrect="off"
            autocapitalize="off"></textarea>`;
}

// attachEventListeners should be called AFTER mainView's innerHTML rewrite
// (so the textarea exists in the DOM). It wires input events + sets initial
// value from state.
export function attachEventListeners() {
  const ta = document.getElementById("editor-textarea");
  if (!ta) return;

  const s = getState();
  ta.value = s.content || "";

  // Lazy-import api to avoid circular deps (main.js imports editor.js too)
  import("../api.js").then(({ api }) => {
    let dirtyNotified = false;
    ta.addEventListener("input", () => {
      const v = ta.value;
      setState({ content: v, dirty: true });
      if (!dirtyNotified) {
        dirtyNotified = true;
        api.setDirty(true);
      }
    });
  });
}