const { contextBridge, ipcRenderer } = require("electron");

contextBridge.exposeInMainWorld("mingo", {
  run: async (source) => {
    return await ipcRenderer.invoke("run-mingo", source);
  },
});
