const state = {
  loaded: false,
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
