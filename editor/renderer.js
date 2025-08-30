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
  async function scheduleDiagnostics() {
    if (diagTimer) clearTimeout(diagTimer);
    diagTimer = setTimeout(runDiagnostics, 250);
  }

  async function runDiagnostics() {
    const src = window.editor.getValue();
    try {
      const res = await window.mingo.diagnostics(src);
      if (!res || !Array.isArray(res.errors)) return;
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
    } catch (_) {
      // ignore
    }
  }

  // Kick initial diagnostics
  scheduleDiagnostics();

  runBtn.addEventListener("click", async () => {
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
  });

  clearBtn.addEventListener("click", () => {
    consoleEl.textContent = "";
  });

  // Placeholder hooks for future features (autocomplete, debugging)
  // monaco.languages.registerCompletionItemProvider('mingo', { ... })
  // Debugging can be integrated using breakpoints overlay + IPC to a debug vm
});
