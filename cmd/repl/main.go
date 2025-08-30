package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"mingo/internal/lexer"
	"mingo/internal/parser"
)

const PROMPT = "mg> "

func main() {
	in := bufio.NewScanner(os.Stdin)
	fmt.Println("Mingo parser REPL. Type code and press Enter. Ctrl+D to exit.")

	var buf strings.Builder
	indent := 0

	for {
		fmt.Print(PROMPT)
		if !in.Scan() {
			break
		}
		line := in.Text()
		buf.WriteString(line)
		buf.WriteString("\n")

		// naive block tracking to wait for closing '}'
		for _, r := range line {
			if r == '{' {
				indent++
			} else if r == '}' {
				if indent > 0 {
					indent--
				}
			}
		}

		if indent > 0 {
			continue
		}

		// attempt to parse the accumulated buffer
		l := lexer.New(buf.String())
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			fmt.Println("parse errors:")
			for _, e := range p.Errors() {
				fmt.Println("  ", e)
			}
		} else {
			fmt.Println(program.String())
		}
		buf.Reset()
	}
}
