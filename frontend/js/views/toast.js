export function showToast(message, opts = {}) {
  const kind = opts.kind || "info"; // "info" | "error"
  const duration = opts.duration ?? (kind === "error" ? null : 2500);
  const container = document.getElementById("toast-container");
  const t = document.createElement("div");
  t.className = `toast toast-${kind}`;
  const icon = kind === "error" ? "⚠" : "✓";
  t.innerHTML = `
    <span class="toast-icon"></span>
    <span class="toast-msg"></span>
    <span class="toast-close">✕</span>
  `;
  t.querySelector(".toast-icon").textContent = icon;
  t.querySelector(".toast-msg").textContent = message;
  t.querySelector(".toast-close").onclick = () => dismissToast(t);
  container.appendChild(t);

  // Auto-dismiss only for non-error toasts
  if (duration !== null) {
    setTimeout(() => fadeOutAndRemove(t), duration);
  }

  // Cap at 5 toasts (drop oldest)
  while (container.children.length > 5) {
    container.firstChild.remove();
  }
}

function fadeOutAndRemove(t) {
  if (!t.parentNode) return;
  t.classList.add("fade-out");
  setTimeout(() => t.remove(), 300);
}

function dismissToast(t) {
  t.remove();
}
