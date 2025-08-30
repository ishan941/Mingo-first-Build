package token

// Type represents the type of a token.
type Type string

// Token represents a lexical token.
type Token struct {
	Type    Type
	Literal string
	Pos     Position
}

// Position indicates the position of a token in the source code.
type Position struct {
	Line   int // 1-based
	Column int // 1-based (rune index)
	Offset int // 0-based byte offset
}

const (
	// Special tokens
	ILLEGAL Type = "ILLEGAL"
	EOF     Type = "EOF"

	// Identifiers + literals
	IDENT Type = "IDENT" // add, foobar, x, y, ...
	INT   Type = "INT"   // 123

	// Keywords
	LET    Type = "LET"
	TRUE   Type = "TRUE"
	FALSE  Type = "FALSE"
	IF     Type = "IF"
	ELSE   Type = "ELSE"
	WHILE  Type = "WHILE"
	FN     Type = "FN"
	RETURN Type = "RETURN"
	PRINT  Type = "PRINT"

	// Operators
	ASSIGN Type = "ASSIGN" // =
	PLUS   Type = "PLUS"   // +
	MINUS  Type = "MINUS"  // -
	ASTER  Type = "ASTER"  // *
	SLASH  Type = "SLASH"  // /

	BANG   Type = "BANG"   // !
	EQ     Type = "EQ"     // ==
	NOT_EQ Type = "NOT_EQ" // !=
	LT     Type = "LT"     // <
	GT     Type = "GT"     // >
	LTE    Type = "LTE"    // <=
	GTE    Type = "GTE"    // >=

	// Delimiters
	COMMA     Type = "COMMA"
	SEMICOLON Type = "SEMICOLON"
	LPAREN    Type = "LPAREN"
	RPAREN    Type = "RPAREN"
	LBRACE    Type = "LBRACE"
	RBRACE    Type = "RBRACE"
)

var keywords = map[string]Type{
	"let":    LET,
	"true":   TRUE,
	"false":  FALSE,
	"if":     IF,
	"else":   ELSE,
	"while":  WHILE,
	"fn":     FN,
	"return": RETURN,
	"print":  PRINT,
}

// LookupIdent returns the token type of the identifier; if it's a keyword, returns the keyword type.
func LookupIdent(ident string) Type {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
