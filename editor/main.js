const { app, BrowserWindow, ipcMain } = require("electron");
const path = require("path");
const fs = require("fs");
const { spawn } = require("child_process");

let mainWindow;

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1200,
    height: 800,
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: true,
      preload: path.join(__dirname, "preload.js"),
    },
  });

  mainWindow.loadFile("index.html");
  if (!app.isPackaged) {
    mainWindow.webContents.openDevTools({ mode: "detach" });
  }
}

ipcMain.handle("run-mingo", async (event, source) => {
  // Build path to VM runner binary (bin/run)
  const repoRoot = path.resolve(__dirname, "..");
  const runner = path.join(
    repoRoot,
    "bin",
    process.platform === "win32" ? "run.exe" : "run"
  );

  if (!fs.existsSync(runner)) {
    return {
      code: -1,
      out: "",
      err: `Mingo runner not found at ${runner}. Build it first (make build).`,
    };
  }

  return new Promise((resolve) => {
    const child = spawn(runner, [], { stdio: ["pipe", "pipe", "pipe"] });
    let out = "";
    let err = "";

    child.stdout.on("data", (d) => {
      out += d.toString();
    });
    child.stderr.on("data", (d) => {
      err += d.toString();
    });

    child.on("close", (code) => {
      resolve({ code, out, err });
    });

    child.on("error", (e) => {
      resolve({
        code: -1,
        out: "",
        err: `Failed to start runner: ${e.message}`,
      });
    });

    child.stdin.write(source);
    child.stdin.end();
  });
});

// Optional: lightweight diagnostics by invoking Go parser via a tiny helper binary would be ideal,
// but we can reuse the runner to just parse and return errors by using a special flag in future.
// For now, keep the channel placeholder so the renderer can call it later without breaking.
ipcMain.handle("mingo-diagnostics", async (_event, source) => {
  const repoRoot = path.resolve(__dirname, "..");
  const diagBin = path.join(
    repoRoot,
    "bin",
    process.platform === "win32" ? "diag.exe" : "diag"
  );
  if (!fs.existsSync(diagBin)) {
    return { errors: [], missing: true };
  }
  return new Promise((resolve) => {
    const child = spawn(diagBin, [], { stdio: ["pipe", "pipe", "pipe"] });
    let out = "";
    let err = "";
    child.stdout.on("data", (d) => (out += d.toString()));
    child.stderr.on("data", (d) => (err += d.toString()));
    child.on("close", () => {
      try {
        const parsed = JSON.parse(out || "[]");
        resolve({ errors: parsed, err });
      } catch (e) {
        resolve({ errors: [], err: String(e) });
      }
    });
    child.on("error", (e) => resolve({ errors: [], err: e.message }));
    child.stdin.write(source);
    child.stdin.end();
  });
});

app.whenReady().then(() => {
  createWindow();

  app.on("activate", function () {
    if (BrowserWindow.getAllWindows().length === 0) createWindow();
  });
});

app.on("window-all-closed", function () {
  if (process.platform !== "darwin") app.quit();
});
