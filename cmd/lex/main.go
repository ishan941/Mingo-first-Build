package main

import (
	"bufio"
	"fmt"
	"os"

	"mingo/internal/lexer"
	"mingo/internal/token"
)

func main() {
	var input string

	if len(os.Args) > 1 {
		// Treat the first argument as a file path
		b, err := os.ReadFile(os.Args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		input = string(b)
	} else {
		// Read from stdin
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				input += scanner.Text() + "\n"
			}
		} else {
			fmt.Println("Usage: lex <file.mg | stdin>")
			os.Exit(2)
		}
	}

	l := lexer.New(input)
	for {
		tok := l.NextToken()
		fmt.Printf("%-10s %-10q @%d:%d (%d)\n", tok.Type, tok.Literal, tok.Pos.Line, tok.Pos.Column, tok.Pos.Offset)
		if tok.Type == token.EOF {
			break
		}
		if tok.Type == token.ILLEGAL {
			os.Exit(3)
		}
	}
}
