package lexer

import (
	"testing"

	"github.com/rhwilr/lemur/token"
)

func TestNextToken(t *testing.T) {
	input := `
let five = 5;
let ten = 10;
const immutable = "test";
f1 f2

let add = fn(x, y) {
	x + y;
};

let result = add(five, ten);
!-*/5;
5 < 10 > 5;

if (5 < 10) {
	return true;
} else {
	return false;
}

while (i < 10) {
}

10 == 10;
10 != 9;
10 <= 9;
10 >= 9;
4 += 1;
4 -= 1;
4 /= 2;
4 *= 2;
&&
||
$
&
|
++
--
"foobar"
"foo bar"
[1, 2];
{"foo": "bar"}
`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.LET, "let"},
		{token.IDENT, "five"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},

		{token.LET, "let"},
		{token.IDENT, "ten"},
		{token.ASSIGN, "="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},

		{token.CONST, "const"},
		{token.IDENT, "immutable"},
		{token.ASSIGN, "="},
		{token.STRING, "test"},
		{token.SEMICOLON, ";"},

		{token.IDENT, "f1"},
		{token.IDENT, "f2"},

		{token.LET, "let"},
		{token.IDENT, "add"},
		{token.ASSIGN, "="},
		{token.FUNCTION, "fn"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.IDENT, "y"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "result"},
		{token.ASSIGN, "="},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.IDENT, "five"},
		{token.COMMA, ","},
		{token.IDENT, "ten"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},

		{token.BANG, "!"},
		{token.MINUS, "-"},
		{token.ASTERISK, "*"},
		{token.SLASH, "/"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.GT, ">"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},

		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.TRUE, "true"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.ELSE, "else"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.FALSE, "false"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},

		{token.WHILE, "while"},
		{token.LPAREN, "("},
		{token.IDENT, "i"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},

		{token.INT, "10"},
		{token.EQ, "=="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.INT, "10"},
		{token.NOT_EQ, "!="},
		{token.INT, "9"},
		{token.SEMICOLON, ";"},
		{token.INT, "10"},
		{token.LT_EQ, "<="},
		{token.INT, "9"},
		{token.SEMICOLON, ";"},
		{token.INT, "10"},
		{token.GT_EQ, ">="},
		{token.INT, "9"},
		{token.SEMICOLON, ";"},

		{token.INT, "4"},
		{token.PLUS_EQUALS, "+="},
		{token.INT, "1"},
		{token.SEMICOLON, ";"},
		{token.INT, "4"},
		{token.MINUS_EQUALS, "-="},
		{token.INT, "1"},
		{token.SEMICOLON, ";"},
		{token.INT, "4"},
		{token.SLASH_EQUALS, "/="},
		{token.INT, "2"},
		{token.SEMICOLON, ";"},
		{token.INT, "4"},
		{token.ASTERISK_EQUALS, "*="},
		{token.INT, "2"},
		{token.SEMICOLON, ";"},

		{token.AND, "&&"},
		{token.OR, "||"},

		{token.ILLEGAL, "$"},
		{token.ILLEGAL, "&"},
		{token.ILLEGAL, "|"},

		{token.PLUS_PLUS, "++"},
		{token.MINUS_MINUS, "--"},

		{token.STRING, "foobar"},
		{token.STRING, "foo bar"},

		{token.LBRACKET, "["},
		{token.INT, "1"},
		{token.COMMA, ","},
		{token.INT, "2"},
		{token.RBRACKET, "]"},
		{token.SEMICOLON, ";"},

		{token.LBRACE, "{"},
		{token.STRING, "foo"},
		{token.COLON, ":"},
		{token.STRING, "bar"},
		{token.RBRACE, "}"},
		{token.EOF, ""},

		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestSimpleComment(t *testing.T) {
	input := `5; // This is a single line comment
// The next line assignes a variable
let number = 1;
// The variable has been assigned,
// but this line is stille a comment.`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "number"},
		{token.ASSIGN, "="},
		{token.INT, "1"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestMultiLineComment(t *testing.T) {
	input := `number/* A comment can start right after source code,
and can span multiple lines.
let nope = "string"; */
let a = 1;
// Single line comment with //

let number /* next we assign the variable */ = 3;`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.IDENT, "number"},
		{token.LET, "let"},
		{token.IDENT, "a"},
		{token.ASSIGN, "="},
		{token.INT, "1"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "number"},
		{token.ASSIGN, "="},
		{token.INT, "3"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}
