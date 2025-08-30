# Mingo Programming Language

This is a tiny educational language with a lexer, parser, compiler to bytecode, and a VM.

Currently implemented: the lexer.

## Layout

- `internal/token`: token types and keyword lookup
- `internal/lexer`: UTF-8 aware lexer with positions
- `cmd/lex`: small CLI to print tokens for a source file or stdin

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

## Test

```sh
go test ./...
```
