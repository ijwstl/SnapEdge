// preload.js
const { contextBridge, ipcRenderer } = require("electron");

contextBridge.exposeInMainWorld("electronAPI", {
  selectImageFile: () => ipcRenderer.invoke("select-image-file"),
  readImageFile: (filePath) => ipcRenderer.invoke("read-image-file", filePath),
});
