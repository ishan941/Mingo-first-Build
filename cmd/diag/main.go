package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"mingo/internal/lexer"
	"mingo/internal/parser"
)

type diag struct {
	Msg    string `json:"msg"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

func main() {
	// Read all from stdin
	src, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read error: %v\n", err)
		os.Exit(1)
	}

	l := lexer.New(string(src))
	p := parser.New(l)
	_ = p.ParseProgram()

	rich := p.RichErrors()
	out := make([]diag, 0, len(rich))
	for _, e := range rich {
		out = append(out, diag{Msg: e.Msg, Line: e.Pos.Line, Column: e.Pos.Column})
	}

	enc := json.NewEncoder(os.Stdout)
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "encode error: %v\n", err)
		os.Exit(2)
	}
}
