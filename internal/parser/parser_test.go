package parser_test

import (
	"testing"

	"mingo/internal/ast"
	"mingo/internal/lexer"
	"mingo/internal/parser"
)

func TestLetStatements(t *testing.T) {
	input := `
let x = 5;
let y = 10;
let foobar = 838383;
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d", len(program.Statements))
	}

	tests := []struct {
		expectedIdentifier string
	}{
		{"x"}, {"y"}, {"foobar"},
	}

	for i, tt := range tests {
		stmt := program.Statements[i]
		if !testLetStatement(t, stmt, tt.expectedIdentifier) {
			return
		}
	}
}

func TestReturnStatements(t *testing.T) {
	input := `
return 5;
return 10;
return 993322;
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d", len(program.Statements))
	}

	for _, stmt := range program.Statements {
		rs, ok := stmt.(*ast.ReturnStatement)
		if !ok {
			t.Fatalf("stmt not *ast.ReturnStatement. got=%T", stmt)
		}
		if rs.TokenLiteral() != "return" {
			t.Fatalf("rs.TokenLiteral not 'return', got %q", rs.TokenLiteral())
		}
	}
}

func TestIfElseWhile(t *testing.T) {
	input := `
if (5 < 10) { return true; } else { return false; }
while (10 > 5) { let x = 1; }
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 2 {
		t.Fatalf("program.Statements expected 2, got %d", len(program.Statements))
	}
}

func TestFunctionAndCall(t *testing.T) {
	input := `
fn add(x, y) { x + y; }
add(1, 2);
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 2 {
		t.Fatalf("program.Statements expected 2, got %d", len(program.Statements))
	}
}

func checkParserErrors(t *testing.T, p *parser.Parser) {
	t.Helper()
	if len(p.Errors()) == 0 {
		return
	}
	for _, e := range p.Errors() {
		t.Log(e)
	}
	t.Fatalf("parser had %d error(s)", len(p.Errors()))
}

func testLetStatement(t *testing.T, s ast.Statement, name string) bool {
	if s.TokenLiteral() != "let" {
		t.Fatalf("s.TokenLiteral not 'let'. got=%q", s.TokenLiteral())
		return false
	}

	letStmt, ok := s.(*ast.LetStatement)
	if !ok {
		t.Fatalf("s not *ast.LetStatement. got=%T", s)
		return false
	}

	if letStmt.Name.Value != name {
		t.Fatalf("letStmt.Name.Value not '%s'. got=%s", name, letStmt.Name.Value)
		return false
	}

	if letStmt.Name.TokenLiteral() != name {
		t.Fatalf("letStmt.Name.TokenLiteral not '%s'. got=%s", name, letStmt.Name.TokenLiteral())
		return false
	}

	return true
}
