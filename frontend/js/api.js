// Wails 在 window.runtime 中暴露 Go 方法
// 路径：window.go.app.Service.Xxx

function call(method, ...args) {
  const fn = window.go?.app?.Service?.[method];
  if (!fn) {
    return Promise.reject(new Error(`method ${method} not bound`));
  }
  return fn(...args);
}

function toResult(res) {
  return {
    ok: res?.ok !== false,
    error: res?.error,
    code: res?.code,
    data: res?.data ?? res,
  };
}

export const api = {
  openFile: (path) => call("OpenFile", path),
  listCommits: () => call("ListCommits"),
  getDiff: (left, right, hideDiff) => call("GetDiff", left, right, hideDiff),
  copySection: (kind) => call("CopySection", kind),
  reload: () => call("Reload"),
  setFontSize: (size) => call("SetFontSize", size),
  getFontSize: () => call("GetFontSize"),
  setFoldComments: (enabled) => call("SetFoldComments", enabled),
  windowChanged: () => call("WindowChanged"),
  // 编辑模式（Task 1/2 新增）
  beginEdit: async () => toResult(await call("BeginEdit")),
  endEdit: async () => toResult(await call("EndEdit")),
  setDirty: async (b) => {
    await call("SetDirty", b);
    return { ok: true };
  },
  save: async (content) => toResult(await call("Save", content)),
  saveOverwrite: async (content) => toResult(await call("SaveOverwrite", content)),
  isDirty: () => call("IsDirty"),
};
