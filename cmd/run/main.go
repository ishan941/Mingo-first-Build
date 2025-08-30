package main

import (
	"bufio"
	"fmt"
	"os"

	"mingo/internal/compiler"
	"mingo/internal/lexer"
	"mingo/internal/parser"
	"mingo/internal/vm"
)

func main() {
	var input string

	if len(os.Args) > 1 {
		b, err := os.ReadFile(os.Args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		input = string(b)
	} else {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				input += scanner.Text() + "\n"
			}
		} else {
			fmt.Println("Usage: mingo-run <file.mg | stdin>")
			os.Exit(2)
		}
	}

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		for _, e := range p.Errors() {
			fmt.Fprintln(os.Stderr, e)
		}
		os.Exit(3)
	}

	comp := compiler.New()
	if err := comp.Compile(program); err != nil {
		fmt.Fprintln(os.Stderr, "compile error:", err)
		os.Exit(4)
	}

	machine := vm.New(comp.Instructions(), comp.Constants())
	if err := machine.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "runtime error:", err)
		os.Exit(5)
	}
}
