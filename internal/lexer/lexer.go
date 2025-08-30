package lexer

import (
	"unicode"
	"unicode/utf8"

	"mingo/internal/token"
)

type Lexer struct {
	input        string
	position     int // byte offset of current char
	readPosition int // byte offset of next char
	ch           rune
	width        int

	line   int
	column int
}

func New(input string) *Lexer {
	l := &Lexer{input: input, line: 1, column: 0}
	l.readRune()
	return l
}

func (l *Lexer) readRune() {
	if l.readPosition >= len(l.input) {
		l.width = 0
		l.ch = 0
		return
	}
	r, w := utf8.DecodeRuneInString(l.input[l.readPosition:])
	l.position = l.readPosition
	l.readPosition += w
	l.width = w

	if r == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}
	l.ch = r
}

func (l *Lexer) peekRune() rune {
	if l.readPosition >= len(l.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.readPosition:])
	return r
}

func (l *Lexer) skipWhitespace() {
	for unicode.IsSpace(l.ch) {
		l.readRune()
	}
}

func (l *Lexer) NextToken() token.Token {
	l.skipWhitespace()

	tok := token.Token{Pos: token.Position{Line: l.line, Column: l.column, Offset: l.position}}

	switch l.ch {
	case '=':
		if l.peekRune() == '=' {
			l.readRune()
			tok.Type = token.EQ
			tok.Literal = "=="
		} else {
			tok.Type = token.ASSIGN
			tok.Literal = "="
		}
	case '+':
		tok.Type = token.PLUS
		tok.Literal = "+"
	case '-':
		tok.Type = token.MINUS
		tok.Literal = "-"
	case '*':
		tok.Type = token.ASTER
		tok.Literal = "*"
	case '/':
		tok.Type = token.SLASH
		tok.Literal = "/"
	case '!':
		if l.peekRune() == '=' {
			l.readRune()
			tok.Type = token.NOT_EQ
			tok.Literal = "!="
		} else {
			tok.Type = token.BANG
			tok.Literal = "!"
		}
	case '<':
		if l.peekRune() == '=' {
			l.readRune()
			tok.Type = token.LTE
			tok.Literal = "<="
		} else {
			tok.Type = token.LT
			tok.Literal = "<"
		}
	case '>':
		if l.peekRune() == '=' {
			l.readRune()
			tok.Type = token.GTE
			tok.Literal = ">="
		} else {
			tok.Type = token.GT
			tok.Literal = ">"
		}
	case ',':
		tok.Type = token.COMMA
		tok.Literal = ","
	case ';':
		tok.Type = token.SEMICOLON
		tok.Literal = ";"
	case '(':
		tok.Type = token.LPAREN
		tok.Literal = "("
	case ')':
		tok.Type = token.RPAREN
		tok.Literal = ")"
	case '{':
		tok.Type = token.LBRACE
		tok.Literal = "{"
	case '}':
		tok.Type = token.RBRACE
		tok.Literal = "}"
	case 0:
		tok.Type = token.EOF
		tok.Literal = ""
	default:
		if isLetter(l.ch) || l.ch == '_' {
			lit := l.readIdentifier()
			tok.Type = token.LookupIdent(lit)
			tok.Literal = lit
			return tok
		} else if unicode.IsDigit(l.ch) {
			lit := l.readNumber()
			tok.Type = token.INT
			tok.Literal = lit
			return tok
		} else {
			tok.Type = token.ILLEGAL
			tok.Literal = string(l.ch)
		}
	}

	l.readRune()
	return tok
}

func (l *Lexer) readIdentifier() string {
	start := l.position
	for isLetter(l.ch) || unicode.IsDigit(l.ch) || l.ch == '_' {
		l.readRune()
	}
	// l.position points to last advanced position; slice using byte indices
	lit := l.input[start:l.position]
	return lit
}

func (l *Lexer) readNumber() string {
	start := l.position
	for unicode.IsDigit(l.ch) {
		l.readRune()
	}
	lit := l.input[start:l.position]
	return lit
}

func isLetter(r rune) bool {
	return unicode.IsLetter(r)
}
