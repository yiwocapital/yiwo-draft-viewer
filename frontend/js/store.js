const state = {
  loaded: false,
  path: "",
  hasGit: false,
  hasFrontmatter: false,
  title: "",
  summary: "",
  content: "",
  charCount: 0,
  commits: [],
  selected: null, // hash or "WORKING"
  diff: { segments: [], charCount: 0, static: false },
  multiSelect: [],
  fontSize: 14,
  foldComments: false,    // "隐藏编辑注释" — strip <!-- --> from diff/static renders
  hideDiff: false,        // "隐藏对比" — render selected commit as plain text, no diff
  search: {
    open: false,
    term: "",       // current search query (trimmed); empty = no highlights
    currentIndex: 0, // 0-based index into the live match list (NOT stored)
  },
};

const subs = new Set();

export function getState() {
  return state;
}

export function setState(patch) {
  Object.assign(state, patch);
  subs.forEach((fn) => fn(state));
}

export function subscribe(fn) {
  subs.add(fn);
  return () => subs.delete(fn);
}
