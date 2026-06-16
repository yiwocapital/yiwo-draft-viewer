export function showToast(message) {
  const container = document.getElementById("toast-container");
  const t = document.createElement("div");
  t.className = "toast";
  t.innerHTML = `<span class="msg"></span><span class="close">✕</span>`;
  t.querySelector(".msg").textContent = message;
  t.querySelector(".close").onclick = () => t.remove();
  container.appendChild(t);
  // Keep max 3 toasts
  while (container.children.length > 3) container.firstChild.remove();
}
