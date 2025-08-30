const { contextBridge, ipcRenderer } = require("electron");

contextBridge.exposeInMainWorld("mingo", {
  run: async (source) => {
    return await ipcRenderer.invoke("run-mingo", source);
  },
  diagnostics: async (source) => {
    return await ipcRenderer.invoke("mingo-diagnostics", source);
  },
  listExamples: async () => {
    return await ipcRenderer.invoke("mingo-list-examples");
  },
  loadExample: async (name) => {
    return await ipcRenderer.invoke("mingo-load-example", name);
  },
  fsList: async () => {
    return await ipcRenderer.invoke("fs-list");
  },
  fsRead: async (relPath) => {
    return await ipcRenderer.invoke("fs-read", relPath);
  },
  fsWrite: async (relPath, content) => {
    return await ipcRenderer.invoke("fs-write", relPath, content);
  },
});
