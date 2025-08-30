// Use Monaco AMD loader shipped in node_modules to avoid bundling
window.require.config({ paths: { vs: "node_modules/monaco-editor/min/vs" } });

window.require(["vs/editor/editor.main"], function () {
  const monaco = window.monaco;

  // Define Mingo language
  monaco.languages.register({ id: "mingo" });
  monaco.languages.setMonarchTokensProvider("mingo", {
    tokenizer: {
      root: [
        [/\b(fn|let|if|else|return|true|false|while|print)\b/, "keyword"],
        [/\d+/, "number"],
        [/\w+/, "identifier"],
        [/==|!=|<=|>=|[=+\-*\/<>!]/, "operator"],
        [/\{|\}|\(|\)|,|;/, "delimiter"],
        [/\s+/, "white"],
      ],
    },
  });

  monaco.languages.setLanguageConfiguration("mingo", {
    comments: { lineComment: "//" },
    brackets: [
      ["{", "}"],
      ["[", "]"],
      ["(", ")"],
    ],
    autoClosingPairs: [
      { open: "{", close: "}" },
      { open: "[", close: "]" },
      { open: "(", close: ")" },
      { open: '"', close: '"' },
      { open: "'", close: "'" },
    ],
  });

  monaco.editor.defineTheme("mingoTheme", {
    base: "vs-dark",
    inherit: true,
    rules: [
      { token: "keyword", foreground: "C586C0" },
      { token: "number", foreground: "B5CEA8" },
      { token: "identifier", foreground: "D4D4D4" },
      { token: "operator", foreground: "D4D4D4" },
      { token: "delimiter", foreground: "D4D4D4" },
    ],
    colors: {},
  });

  const initial =
    localStorage.getItem("mingo.source") ?? "let x = 10;\nprint(x);";
  window.editor = monaco.editor.create(document.getElementById("editor"), {
    value: initial,
    language: "mingo",
    theme: "mingoTheme",
    automaticLayout: true,
    minimap: { enabled: false },
  });

  const runBtn = document.getElementById("runBtn");
  const consoleEl = document.getElementById("console");
  const clearBtn = document.getElementById("clearBtn");
  const statusEl = document.getElementById("status");
  const examplesEl = document.getElementById("examples");
  const formatBtn = document.getElementById("formatBtn");
  const fileTreeEl = document.getElementById("fileTree");
  let currentFile = null;

  // Cmd/Ctrl+Enter to Run
  window.editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter, () => {
    runBtn.click();
  });

  // Completion provider: keywords, snippets, and identifiers from buffer
  const keywords = [
    "fn",
    "let",
    "if",
    "else",
    "return",
    "true",
    "false",
    "while",
    "print",
  ];
  const keywordSet = new Set(keywords);
  monaco.languages.registerCompletionItemProvider("mingo", {
    triggerCharacters: [" ", "(", ")", ",", ";", "\n"],
    provideCompletionItems(model, position) {
      const text = model.getValue();
      const idRegex = /\b[a-zA-Z_][a-zA-Z0-9_]*\b/g;
      const identifiers = new Set();
      let m;
      while ((m = idRegex.exec(text))) {
        const w = m[0];
        if (!keywordSet.has(w)) identifiers.add(w);
      }

      /** @type {import('monaco-editor').languages.CompletionItem[]} */
      const suggestions = [];

      // Keywords
      for (const k of keywords) {
        suggestions.push({
          label: k,
          kind: monaco.languages.CompletionItemKind.Keyword,
          insertText: k,
        });
      }

      // Snippets
      suggestions.push(
        {
          label: "fn snippet",
          kind: monaco.languages.CompletionItemKind.Snippet,
          detail: "function",
          insertTextRules:
            monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
          insertText: "fn ${1:name}(${2:args}) {\n\t$0\n}",
        },
        {
          label: "if snippet",
          kind: monaco.languages.CompletionItemKind.Snippet,
          detail: "if/else",
          insertTextRules:
            monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
          insertText: "if (${1:cond}) {\n\t$0\n} else {\n\t\n}",
        },
        {
          label: "while snippet",
          kind: monaco.languages.CompletionItemKind.Snippet,
          detail: "while loop",
          insertTextRules:
            monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
          insertText: "while (${1:cond}) {\n\t$0\n}",
        },
        {
          label: "print",
          kind: monaco.languages.CompletionItemKind.Function,
          insertText: "print(${1:expr});",
          insertTextRules:
            monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
        }
      );

      // Buffer identifiers
      for (const id of identifiers) {
        suggestions.push({
          label: id,
          kind: monaco.languages.CompletionItemKind.Variable,
          insertText: id,
        });
      }

      return { suggestions };
    },
  });

  function appendConsole(text) {
    consoleEl.textContent += text;
    consoleEl.scrollTop = consoleEl.scrollHeight;
  }

  // Persist source between sessions
  window.editor.onDidChangeModelContent(() => {
    localStorage.setItem("mingo.source", window.editor.getValue());
    scheduleDiagnostics();
  });

  let diagTimer = null;
  let isRunning = false;
  async function scheduleDiagnostics() {
    if (diagTimer) clearTimeout(diagTimer);
    diagTimer = setTimeout(runDiagnostics, 250);
  }

  async function runDiagnostics() {
    const src = window.editor.getValue();
    try {
      const res = await window.mingo.diagnostics(src);
      if (!res || res.missing || !Array.isArray(res.errors)) return;
      const model = window.editor.getModel();
      const markers = res.errors.map((e) => ({
        severity: monaco.MarkerSeverity.Error,
        message: e.msg || e.Msg || "Error",
        startLineNumber: e.line || e.Line || 1,
        startColumn: e.column || e.Column || 1,
        endLineNumber: e.line || e.Line || 1,
        endColumn: (e.column || e.Column || 1) + 1,
      }));
      monaco.editor.setModelMarkers(model, "mingo", markers);
      if (!isRunning) {
        if (markers.length > 0) {
          statusEl.textContent = `Issues: ${markers.length}`;
          statusEl.style.color = "#ff7777";
        } else {
          statusEl.textContent = "Ready";
          statusEl.style.color = "";
        }
      }
    } catch (_) {
      // ignore
    }
  }

  // Kick initial diagnostics
  scheduleDiagnostics();

  // Populate examples dropdown
  (async function loadExamples() {
    try {
      const list = await window.mingo.listExamples();
      examplesEl.innerHTML = "";
      const def = document.createElement("option");
      def.textContent = "Examples…";
      def.selected = true;
      def.disabled = true;
      examplesEl.appendChild(def);
      for (const name of list) {
        const opt = document.createElement("option");
        opt.value = name;
        opt.textContent = name;
        examplesEl.appendChild(opt);
      }
    } catch (e) {
      examplesEl.innerHTML = "<option>Error loading examples</option>";
    }
  })();

  examplesEl.addEventListener("change", async () => {
    const name = examplesEl.value;
    if (!name) return;
    const res = await window.mingo.loadExample(name);
    if (res && res.ok) {
      window.editor.setValue(res.content);
    } else if (res && res.error) {
      appendConsole(res.error + "\n");
    }
  });

  // File explorer
  function buildTree(items) {
    // items: [{type:'dir'|'file', name, path}]
    const root = {};
    for (const it of items) {
      const parts = it.path.split("/");
      let node = root;
      for (let i = 0; i < parts.length; i++) {
        const part = parts[i];
        const isLast = i === parts.length - 1;
        if (isLast) {
          // last segment: directory node or file leaf (null)
          node[part] = node[part] ?? (it.type === "dir" ? {} : null);
        } else {
          // descend into directory segment
          node[part] = node[part] ?? {};
          node = node[part];
        }
      }
    }
    return root;
  }

  function renderTree(node, base = "") {
    const ul = document.createElement("ul");
    const entries = Object.keys(node).sort((a, b) => {
      const ad = node[a] && typeof node[a] === "object";
      const bd = node[b] && typeof node[b] === "object";
      if (ad !== bd) return ad ? -1 : 1;
      return a.localeCompare(b);
    });
    for (const name of entries) {
      const child = node[name];
      const li = document.createElement("li");
      const full = base ? base + "/" + name : name;
      if (child && typeof child === "object") {
        const twist = document.createElement("span");
        twist.className = "twisty";
        twist.textContent = "▾";
        const span = document.createElement("span");
        span.className = "dir dim";
        span.textContent = name;
        const header = document.createElement("div");
        header.appendChild(twist);
        header.appendChild(span);
        header.style.display = "flex";
        header.style.alignItems = "center";
        li.appendChild(header);
        const childUl = renderTree(child, full);
        li.appendChild(childUl);
        let collapsed = false;
        header.addEventListener("click", () => {
          collapsed = !collapsed;
          childUl.style.display = collapsed ? "none" : "";
          twist.textContent = collapsed ? "▸" : "▾";
        });
      } else {
        const span = document.createElement("span");
        span.className = "file";
        span.textContent = name;
        span.addEventListener("click", async () => {
          const res = await window.mingo.fsRead(full);
          if (res && res.ok) {
            window.editor.setValue(res.content);
            currentFile = full;
            statusEl.textContent = full;
            statusEl.style.color = "";
            // highlight selection
            fileTreeEl
              .querySelectorAll(".file.selected")
              .forEach((el) => el.classList.remove("selected"));
            span.classList.add("selected");
          } else if (res && res.error) {
            appendConsole(res.error + "\n");
          }
        });
        li.appendChild(span);
      }
      ul.appendChild(li);
    }
    return ul;
  }

  async function loadFileTree() {
    fileTreeEl.textContent = "Loading files…";
    try {
      const items = await window.mingo.fsList();
      fileTreeEl.textContent = "";
      fileTreeEl.appendChild(renderTree(buildTree(items)));
    } catch (e) {
      fileTreeEl.textContent = "Failed to load files";
    }
  }

  loadFileTree();

  // Save: Cmd/Ctrl+S
  window.editor.addCommand(
    monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS,
    async () => {
      if (!currentFile) {
        statusEl.textContent = "No file selected";
        statusEl.style.color = "#ff7777";
        return;
      }
      const res = await window.mingo.fsWrite(
        currentFile,
        window.editor.getValue()
      );
      if (res && res.ok) {
        statusEl.textContent = `Saved ${currentFile}`;
        statusEl.style.color = "#77ff77";
        setTimeout(() => {
          statusEl.textContent = currentFile;
          statusEl.style.color = "";
        }, 1200);
      } else if (res && res.error) {
        statusEl.textContent = res.error;
        statusEl.style.color = "#ff7777";
      }
    }
  );

  runBtn.addEventListener("click", async () => {
    isRunning = true;
    statusEl.textContent = "Running...";
    consoleEl.textContent = "";
    const source = window.editor.getValue();
    try {
      const result = await window.mingo.run(source);
      if (result.out) appendConsole(result.out);
      if (result.err) appendConsole(result.err);
      statusEl.textContent = `Exit ${result.code}`;
    } catch (e) {
      appendConsole(String(e));
      statusEl.textContent = "Error";
    }
    isRunning = false;
    // Refresh diagnostics after run
    scheduleDiagnostics();
  });

  clearBtn.addEventListener("click", () => {
    consoleEl.textContent = "";
  });

  // Simple formatter: normalizes spaces, ensures semicolons/newlines and braces indentation
  function formatMingo(src) {
    const lines = [];
    const tokens = src
      .replace(/\t/g, "  ")
      .replace(/\s+/g, (m) => (m.includes("\n") ? "\n" : " "))
      .split(/(\{|\}|\(|\)|;|,)/);

    let out = [];
    let indent = 0;
    const pushLine = (s = "") =>
      lines.push("  ".repeat(Math.max(indent, 0)) + s.trim());

    for (let i = 0; i < tokens.length; i++) {
      const t = tokens[i];
      if (!t) continue;
      if (t === "{") {
        const current = out.join(" ").trim();
        if (current) pushLine(current + " ");
        pushLine("{");
        out = [];
        indent++;
      } else if (t === "}") {
        const current = out.join(" ").trim();
        if (current) {
          pushLine(current + (current.endsWith(";") ? "" : ";"));
          out = [];
        }
        indent--;
        pushLine("}");
      } else if (t === ";") {
        const current = out.join(" ").trim();
        pushLine(current + ";");
        out = [];
      } else if (t === ",") {
        out.push(",");
      } else if (t.match(/^\s*\n\s*$/)) {
        // collapse stray newlines by closing current buffer
        const current = out.join(" ").trim();
        if (current) {
          pushLine(current);
          out = [];
        }
      } else {
        out.push(t.trim());
      }
    }
    const tail = out.join(" ").trim();
    if (tail) pushLine(tail + (tail.endsWith(";") ? "" : ";"));
    return lines.join("\n").replace(/\n{3,}/g, "\n\n");
  }

  function doFormat() {
    const src = window.editor.getValue();
    const formatted = formatMingo(src);
    window.editor.setValue(formatted);
  }

  formatBtn.addEventListener("click", doFormat);
  window.editor.addCommand(
    monaco.KeyMod.CtrlCmd | monaco.KeyMod.Shift | monaco.KeyCode.KeyF,
    doFormat
  );

  // Hover provider to show diagnostic messages on hover
  monaco.languages.registerHoverProvider("mingo", {
    provideHover(model, position) {
      const markers = monaco.editor.getModelMarkers({ resource: model.uri });
      for (const m of markers) {
        const inLine =
          position.lineNumber >= m.startLineNumber &&
          position.lineNumber <= m.endLineNumber;
        const inCol =
          position.column >= m.startColumn && position.column <= m.endColumn;
        if (inLine && inCol) {
          return {
            range: new monaco.Range(
              m.startLineNumber,
              m.startColumn,
              m.endLineNumber,
              m.endColumn
            ),
            contents: [{ value: `Error: ${m.message}` }],
          };
        }
      }
      return null;
    },
  });

  // Quick fixes: insert missing semicolons when diagnostic suggests it
  monaco.languages.registerCodeActionProvider("mingo", {
    providedCodeActionKinds: [monaco.languages.CodeActionKind.QuickFix],
    provideCodeActions(model, range, context) {
      const actions = [];
      for (const m of context.markers || []) {
        const msg = (m.message || "").toUpperCase();
        if (
          msg.includes("EXPECTED NEXT TOKEN TO BE") &&
          msg.includes("SEMICOLON")
        ) {
          const line = m.endLineNumber || m.startLineNumber || 1;
          const col = model.getLineMaxColumn(line);
          actions.push({
            title: "Insert ; at end of line",
            kind: monaco.languages.CodeActionKind.QuickFix,
            edit: {
              edits: [
                {
                  resource: model.uri,
                  edit: {
                    range: new monaco.Range(line, col, line, col),
                    text: ";",
                  },
                },
              ],
            },
            diagnostics: [m],
            isPreferred: true,
          });
        }
      }
      return { actions, dispose: () => {} };
    },
  });

  // Placeholder hooks for future features (autocomplete, debugging)
  // monaco.languages.registerCompletionItemProvider('mingo', { ... })
  // Debugging can be integrated using breakpoints overlay + IPC to a debug vm
});
