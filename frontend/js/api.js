// Wails 在 window.runtime 中暴露 Go 方法
// 路径：window.go.app.Service.Xxx

function call(method, ...args) {
  const fn = window.go?.app?.Service?.[method];
  if (!fn) {
    return Promise.reject(new Error(`method ${method} not bound`));
  }
  return fn(...args);
}

export const api = {
  openFile: (path) => call("OpenFile", path),
  listCommits: () => call("ListCommits"),
  getDiff: (left, right) => call("GetDiff", left, right),
  copySection: (kind) => call("CopySection", kind),
  reload: () => call("Reload"),
  setFontSize: (size) => call("SetFontSize", size),
  getFontSize: () => call("GetFontSize"),
  windowChanged: () => call("WindowChanged"),
};
