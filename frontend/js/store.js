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
  foldComments: false,    // NEW
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
