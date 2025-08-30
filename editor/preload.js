const { contextBridge, ipcRenderer } = require("electron");

contextBridge.exposeInMainWorld("mingo", {
  run: async (source) => {
    return await ipcRenderer.invoke("run-mingo", source);
  },
  diagnostics: async (source) => {
    return await ipcRenderer.invoke("mingo-diagnostics", source);
  },
});
