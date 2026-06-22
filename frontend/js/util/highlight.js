// DOM-based search highlighter. Operates on a root element containing
// `.diff-content` blocks (the diff/static rows mainView renders). Walks the
// text nodes inside those blocks, splits them at match boundaries, and wraps
// the match ranges in <span class="search-match"> (with an extra
// `search-match-current` class on the active match).
//
// Implementation uses the Range API + extractContents/insertNode. This is
// cleaner and more correct than manual splitText + sibling walking: Range
// automatically handles boundaries between text nodes and across inline
// element boundaries (like .diff-ins/.diff-del), so a single match spanning
// "added word" in green and "deleted" in red wraps correctly in one span.
//
// `removeAll` unwraps any existing .search-match spans. Split text nodes
// created by Range are restored as-is (replaceWith childNodes); display is
// identical.

function findMatches(flatText, term) {
  if (!term) return [];
  const lowerText = flatText.toLowerCase();
  const lowerTerm = term.toLowerCase();
  if (lowerTerm.length === 0) return [];
  const out = [];
  let i = 0;
  while ((i = lowerText.indexOf(lowerTerm, i)) !== -1) {
    out.push([i, i + term.length]);
    i += term.length || 1;
  }
  return out;
}

// Walk all text nodes inside `.diff-content` blocks; return both an array of
// {node, start, end} segments AND the concatenated flatText. start/end are
// offsets into flatText.
function walkTextNodes(root) {
  const blocks = root.querySelectorAll(".diff-content");
  const segments = [];
  let offset = 0;
  let flatText = "";
  for (const block of blocks) {
    const walker = document.createTreeWalker(block, NodeFilter.SHOW_TEXT, null);
    let node;
    while ((node = walker.nextNode())) {
      const len = node.nodeValue.length;
      if (len === 0) continue;
      segments.push({ node, start: offset, end: offset + len });
      flatText += node.nodeValue;
      offset += len;
    }
  }
  return { segments, flatText };
}

// Build a Range that covers [start, end) in flatText. Returns null if the
// offsets can't be mapped (e.g., past the end of available text).
function rangeFromOffsets(root, segments, start, end) {
  // Locate the start text node + offset within it.
  let startNode = null, startOffset = 0;
  let endNode = null, endOffset = 0;
  let acc = 0;
  for (const s of segments) {
    const len = s.node.nodeValue.length;
    if (startNode === null && start <= acc + len) {
      startNode = s.node;
      startOffset = start - acc;
    }
    if (endNode === null && end <= acc + len) {
      endNode = s.node;
      endOffset = end - acc;
    }
    acc += len;
    if (startNode && endNode) break;
  }
  if (!startNode || !endNode) return null;

  const range = document.createRange();
  try {
    range.setStart(startNode, startOffset);
    range.setEnd(endNode, endOffset);
  } catch (_) {
    return null;
  }
  return range;
}

export function removeAll(root) {
  const spans = root.querySelectorAll(".search-match");
  for (const s of spans) {
    s.replaceWith(...s.childNodes);
  }
}

export function applyTerm(root, term, currentIndex) {
  removeAll(root);
  if (!term) {
    return { count: 0, currentIndex: 0, firstNode: null };
  }

  const initial = walkTextNodes(root);
  if (initial.segments.length === 0) {
    return { count: 0, currentIndex: 0, firstNode: null };
  }

  const matches = findMatches(initial.flatText, term);
  if (matches.length === 0) {
    return { count: 0, currentIndex: 0, firstNode: null };
  }

  const clampedIndex = Math.min(Math.max(0, currentIndex), matches.length - 1);

  // Process matches from last to first. Earlier (lower-index) matches may
  // be in nodes that haven't been touched yet, but later matches have been
  // processed first and haven't affected earlier positions (non-overlapping
  // substring matches), so rangeFromOffsets with re-walked segments is
  // always correct.
  for (let m = matches.length - 1; m >= 0; m--) {
    const [start, end] = matches[m];
    const isCurrent = m === clampedIndex;

    // Re-walk each iteration: previous (later) matches may have split or
    // moved text nodes via extractContents, so segment references are stale.
    const fresh = walkTextNodes(root);
    const range = rangeFromOffsets(root, fresh.segments, start, end);
    if (!range) continue;

    const span = document.createElement("span");
    span.className = isCurrent ? "search-match search-match-current" : "search-match";

    // extractContents pulls the range out of the DOM as a DocumentFragment
    // (splitting text nodes at boundaries as needed). We then place that
    // fragment into our span and insert the span at the range's start
    // position. Handles inline-element boundaries (e.g. .diff-ins) without
    // the constraint errors of range.surroundContents.
    const frag = range.extractContents();
    span.appendChild(frag);
    range.insertNode(span);
  }

  const firstCurrent = root.querySelector(".search-match-current");
  return { count: matches.length, currentIndex: clampedIndex, firstNode: firstCurrent };
}
