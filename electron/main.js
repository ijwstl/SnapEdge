const { app, BrowserWindow, ipcMain, dialog } = require("electron");
const fs = require("fs");
const path = require("path");

function createWindow() {
  const win = new BrowserWindow({
    width: 800,
    height: 600,
    webPreferences: {
      preload: path.join(__dirname, "preload.js"),
      contextIsolation: true, // 必须开启，确保使用 contextBridge
      enableRemoteModule: false,
      nodeIntegration: false, // 必须关闭，防止渲染进程直接使用 Node.js API
    },
  });

  // 加载 Vue 应用
  console.log(process.env.NODE_ENV);
  if (process.env.NODE_ENV == "production") {
    // 生产环境逻辑
    win.loadFile(path.join(__dirname, "../dist/index.html"));
  } else {
    // 开发环境逻辑
    win.loadURL("http://localhost:8080");
  }
}

// 监听渲染进程的文件加载请求
// 处理文件选择对话框
ipcMain.handle("select-image-file", async () => {
  const result = await dialog.showOpenDialog({
    properties: ["openFile"],
    filters: [{ name: "Images", extensions: ["jpg", "jpeg", "png"] }],
  });
  return result.filePaths[0] || null;
});

// 读取文件内容并返回
ipcMain.handle("read-image-file", async (event, filePath) => {
  return new Promise((resolve, reject) => {
    fs.readFile(filePath, (err, data) => {
      if (err) {
        return reject(err);
      }
      // 将文件内容转换为 base64
      const base64Data = Buffer.from(data).toString("base64");
      resolve(`data:image/jpeg;base64,${base64Data}`); // 适配图片类型
    });
  });
});

app.whenReady().then(() => {
  createWindow();

  app.on("activate", () => {
    if (BrowserWindow.getAllWindows().length === 0) createWindow();
  });
});

app.on("window-all-closed", () => {
  if (process.platform !== "darwin") app.quit();
});
