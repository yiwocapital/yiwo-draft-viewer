const htmlCommentRegex = /<!--[\s\S]*?-->/g;

export function escapeHtml(s) {
  return s.replace(/[&<>"']/g, (m) => ({
    "&": "&amp;",
    "<": "&lt;",
    ">": "&gt;",
    '"': "&quot;",
    "'": "&#39;",
  }[m]));
}

export function stripComments(text) {
  return text.replace(htmlCommentRegex, "");
}
