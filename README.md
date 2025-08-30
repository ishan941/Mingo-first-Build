# Mingo Programming Language

This is a tiny educational language with a lexer, parser, bytecode compiler, and a stack-based VM, plus REPLs.

## Layout

- `internal/token`: token types and keyword lookup
- `internal/lexer`: UTF-8 aware lexer with positions
- `internal/ast`: AST nodes for statements and expressions
- `internal/parser`: Pratt parser and recursive-descent for statements
- `internal/code`: bytecode instruction set and encoder/decoder
- `internal/compiler`: AST -> bytecode compiler, symbol table
- `internal/object`: runtime objects (int, bool, null, compiled function)
- `internal/vm`: stack-based virtual machine
- `cmd/lex`: token dump CLI
- `cmd/repl`: parser REPL (prints AST)
- `cmd/run`: compile+run a program
- `cmd/vmrepl`: VM REPL that preserves state and echoes results

## Try it

Build the lexer CLI and run it over a file:

```sh
go build -o bin/lex ./cmd/lex
./bin/lex path/to/file.mg
```

Or pipe source:

```sh
echo 'let x = 10; print(x);' | ./bin/lex
```

Build parser REPL (AST printer):

```sh
go build -o bin/repl ./cmd/repl
./bin/repl
```

Build runner (compiler+VM):

```sh
go build -o bin/run ./cmd/run
printf 'print(1+2);\nlet x = 10; print(x);\n' | ./bin/run
```

Build VM REPL (stateful, echoes expression results):

```sh
go build -o bin/vmrepl ./cmd/vmrepl
./bin/vmrepl
# examples:
# 1+2;
# let x = 5;
# x;
# x = x + 7;
# x;
```

## Test

```sh
go test ./...
```

## Editor (Electron + Monaco)

A simple desktop editor lives in `editor/` with syntax highlighting and a Run button that executes your code through the Go VM.

Prerequisites:

- Go toolchain installed
- Node.js 18+ and npm

Steps:

```sh
make build                # builds bin/run used by the editor
cd editor
npm install
npm start
```

Usage:

- Type code in the editor; press the Run ▶ button or Cmd/Ctrl+Enter to execute.
- Output appears in the console panel; use Clear Console to reset.
- The editor persists your last program locally between sessions.
