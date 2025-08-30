package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"mingo/internal/code"
	"mingo/internal/compiler"
	"mingo/internal/lexer"
	"mingo/internal/object"
	"mingo/internal/parser"
	"mingo/internal/vm"
)

const PROMPT = "vm> "

func main() {
	fmt.Println("Mingo VM REPL. Type code; Ctrl+D to exit.")

	in := bufio.NewScanner(os.Stdin)

	sym := compiler.NewSymbolTable()
	comp := compiler.NewWithState(sym, nil)
	globals := make([]object.Object, vm.GlobalsSize)

	// naive multi-line buffer for blocks
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

		for _, r := range line {
			if r == '{' {
				indent++
			}
			if r == '}' && indent > 0 {
				indent--
			}
		}
		if indent > 0 {
			continue
		}

		source := buf.String()
		buf.Reset()

		l := lexer.New(source)
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			fmt.Println("parse errors:")
			for _, e := range p.Errors() {
				fmt.Println("  ", e)
			}
			continue
		}

		// re-use compiler state
		comp = compiler.NewWithState(sym, comp.Constants())
		if err := comp.Compile(program); err != nil {
			fmt.Println("compile error:", err)
			continue
		}

		// Echo expressions: replace OpPop with OpPrint for top-level statements
		patched := append([]byte(nil), comp.Instructions()...)
		for i := 0; i < len(patched); {
			op := code.Opcode(patched[i])
			def, _ := code.Lookup(op)
			if op == code.OpPop {
				patched[i] = byte(code.OpPrint)
			}
			// advance index by 1 + operand widths
			read := 0
			if def != nil {
				for _, w := range def.OperandWidths {
					read += w
				}
			}
			i += 1 + read
		}

		// re-use VM globals across iterations
		machine := vm.NewWithGlobals(patched, comp.Constants(), globals)
		if err := machine.Run(); err != nil {
			fmt.Println("runtime error:", err)
			continue
		}
	}
}
