package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	// Special types
	ILLEGAL = "ILLEGAL" // Used when we encounter an unknown token
	EOF     = "EOF"     // Used to signal the end of a file to the parser

	// Indentifiers and literals
	IDENT = "IDENT" // Identifier like variable names
	INT   = "INT"   // Integers 1, 2,3, 42...

	// Assignments
	ASSIGN          = "="
	PLUS_EQUALS     = "+="
	MINUS_EQUALS    = "-="
	SLASH_EQUALS    = "/="
	ASTERISK_EQUALS = "*="

	// Postfix
	MINUS_MINUS = "--"
	PLUS_PLUS = "++"

	// Operators
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"

	LT     = "<"
	GT     = ">"
	EQ     = "=="
	NOT_EQ = "!="

	// Delimiters
	COMMA     = ","
	SEMICOLON = ";"
	COLON     = ":"

	LPAREN   = "("
	RPAREN   = ")"
	LBRACE   = "{"
	RBRACE   = "}"
	LBRACKET = "["
	RBRACKET = "]"

	// Keywords
	FUNCTION = "FUNCTION"
	LET      = "LET"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"

	// String, Array and Hash
	STRING = "STRING"
)

var keywords = map[string]TokenType{
	"fn":     FUNCTION,
	"let":    LET,
	"true":   TRUE,
	"false":  FALSE,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
}

// LookupIdent checks, if the passed identifiers is reserved words. If that is
// the case, it will return the proper TokenType for that identifier.
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
