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

// File explorer: list files and read/save
function listFilesRecursive(dir, baseDir, out, maxDepth = 5) {
  if (maxDepth < 0) return;
  let ents = [];
  try {
    ents = fs.readdirSync(dir, { withFileTypes: true });
  } catch {
    return;
  }
  for (const e of ents) {
    if (e.name === ".git" || e.name === "node_modules" || e.name === "bin")
      continue;
    if (e.name.startsWith(".")) continue; // hide dotfiles (e.g., .gitignore)
    const abs = path.join(dir, e.name);
    const rel = path.relative(baseDir, abs);
    if (e.isDirectory()) {
      out.push({ type: "dir", name: e.name, path: rel });
      listFilesRecursive(abs, baseDir, out, maxDepth - 1);
    } else {
      out.push({ type: "file", name: e.name, path: rel });
    }
  }
}

ipcMain.handle("fs-list", async () => {
  const repoRoot = path.resolve(__dirname, "..");
  const items = [];
  listFilesRecursive(repoRoot, repoRoot, items, 6);
  return items;
});

ipcMain.handle("fs-read", async (_event, relPath) => {
  const repoRoot = path.resolve(__dirname, "..");
  const abs = path.normalize(path.join(repoRoot, relPath));
  if (!abs.startsWith(repoRoot)) return { ok: false, error: "Invalid path" };
  try {
    const content = fs.readFileSync(abs, "utf8");
    return { ok: true, content };
  } catch (e) {
    return { ok: false, error: e.message };
  }
});

ipcMain.handle("fs-write", async (_event, relPath, content) => {
  const repoRoot = path.resolve(__dirname, "..");
  const abs = path.normalize(path.join(repoRoot, relPath));
  if (!abs.startsWith(repoRoot)) return { ok: false, error: "Invalid path" };
  try {
    fs.mkdirSync(path.dirname(abs), { recursive: true });
    fs.writeFileSync(abs, content, "utf8");
    return { ok: true };
  } catch (e) {
    return { ok: false, error: e.message };
  }
});

// List available example files (.mg) from repo root and examples/ directory
ipcMain.handle("mingo-list-examples", async () => {
  const repoRoot = path.resolve(__dirname, "..");
  const candidates = [];
  try {
    const rootFiles = fs.readdirSync(repoRoot, { withFileTypes: true });
    for (const e of rootFiles) {
      if (e.isFile() && e.name.endsWith(".mg"))
        candidates.push({ name: e.name, dir: repoRoot });
    }
  } catch {}
  try {
    const exDir = path.join(repoRoot, "examples");
    const exFiles = fs.readdirSync(exDir, { withFileTypes: true });
    for (const e of exFiles) {
      if (e.isFile() && e.name.endsWith(".mg"))
        candidates.push({ name: `examples/${e.name}`, dir: exDir });
    }
  } catch {}
  // Return names only; loading will resolve again
  return candidates.map((c) => c.name);
});

// Load example file content by name (basename or examples/relative)
ipcMain.handle("mingo-load-example", async (_event, name) => {
  const repoRoot = path.resolve(__dirname, "..");
  const attempts = [];
  if (name.startsWith("examples/")) {
    attempts.push(path.join(repoRoot, name));
  } else {
    attempts.push(path.join(repoRoot, name));
    attempts.push(path.join(repoRoot, "examples", name));
  }
  for (const pth of attempts) {
    try {
      const stat = fs.statSync(pth);
      if (stat.isFile()) {
        const content = fs.readFileSync(pth, "utf8");
        return { ok: true, content };
      }
    } catch {}
  }
  return { ok: false, error: `Example not found: ${name}` };
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
