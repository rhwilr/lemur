package lexer

import (
	"github.com/rhwilr/lemur/token"
)

// The Lexer is used to parse source code and turn it into tokens
type Lexer struct {
	input        []rune
	
	line         int  // current line in input
	column       int  // current column in input
	readColumn       int  // current column in input

	position     int  // current position in input
	readPosition int  // current reading position in input
	ch           rune // current char under examination
}

// New will return a new instance of a Lexer
func New(input string) *Lexer {
	l := &Lexer{
		input:  []rune(input),
		line:   1,
		column: 0,
	}

	l.readChar()

	return l
}

// NextToken will try to parse one ore more characters and return the
// corresponding token
func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()
	
	l.column = l.readColumn

	if l.ch == '/' && l.peekChar() == '/' {
		l.skipSinglLineComments()
		return l.NextToken()
	}

	if l.ch == '/' && l.peekChar() == '*' {
		l.skipMultiLineComments()
		return l.NextToken()
	}

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.EQ, string(ch)+string(l.ch))
		} else {
			tok = l.newTokenFromRune(token.ASSIGN, l.ch)
		}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.NOT_EQ, string(ch)+string(l.ch))
		} else {
			tok = l.newTokenFromRune(token.BANG, l.ch)
		}
	case '+':
		if l.peekChar() == '+' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.PLUS_PLUS, string(ch)+string(l.ch))
		} else if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.PLUS_EQUALS, string(ch)+string(l.ch))
		} else {
			tok = l.newTokenFromRune(token.PLUS, l.ch)
		}
	case '-':
		if l.peekChar() == '-' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.MINUS_MINUS, string(ch)+string(l.ch))
		} else if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.MINUS_EQUALS, string(ch)+string(l.ch))
		} else {
			tok = l.newTokenFromRune(token.MINUS, l.ch)
		}
	case '(':
		tok = l.newTokenFromRune(token.LPAREN, l.ch)
	case ')':
		tok = l.newTokenFromRune(token.RPAREN, l.ch)
	case '{':
		tok = l.newTokenFromRune(token.LBRACE, l.ch)
	case '}':
		tok = l.newTokenFromRune(token.RBRACE, l.ch)
	case '[':
		tok = l.newTokenFromRune(token.LBRACKET, l.ch)
	case ']':
		tok = l.newTokenFromRune(token.RBRACKET, l.ch)
	case '/':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.SLASH_EQUALS, string(ch)+string(l.ch))
		} else {
			tok = l.newTokenFromRune(token.SLASH, l.ch)
		}
	case '*':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.ASTERISK_EQUALS, string(ch)+string(l.ch))
		} else {
			tok = l.newTokenFromRune(token.ASTERISK, l.ch)
		}
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.LT_EQ, string(ch)+string(l.ch))
		} else {
			tok = l.newTokenFromRune(token.LT, l.ch)
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.GT_EQ, string(ch)+string(l.ch))
		} else {
			tok = l.newTokenFromRune(token.GT, l.ch)
		}
	case '&':
		if l.peekChar() == '&' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.AND, string(ch)+string(l.ch))
		} else {
			tok = l.newTokenFromRune(token.ILLEGAL, l.ch)
		}
	case '|':
		if l.peekChar() == '|' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(token.OR, string(ch)+string(l.ch))
		} else {
			tok = l.newTokenFromRune(token.ILLEGAL, l.ch)
		}
	case ',':
		tok = l.newTokenFromRune(token.COMMA, l.ch)
	case ';':
		tok = l.newTokenFromRune(token.SEMICOLON, l.ch)
	case ':':
		tok = l.newTokenFromRune(token.COLON, l.ch)
	case '"':
		tok = l.newToken(token.STRING, l.readStringLiteral())
	case 0:
		tok = l.newToken(token.EOF, "")
	default:
		if isLetter(l.ch) {
			literal := l.readItentifier()
	
			return l.newToken(token.LookupIdent(literal), literal)
		} else if isDigit(l.ch) {
			return l.newToken(token.INT, l.readNumber())
		}

		tok = l.newTokenFromRune(token.ILLEGAL, l.ch)
	}

	l.readChar()

	return tok
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++

	l.readColumn++

	if l.ch == '\n' || l.ch == '\r' {
		l.line++
		l.column = 0
		l.readColumn = 0
	}
}

func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) skipSinglLineComments() {
	// consume characters until we encounter a newline or the end of the file
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}

	l.skipWhitespace()
}

func (l *Lexer) skipMultiLineComments() {
	// consume characters until we encounter the end of the comment or EOF
	found := false

	for !found {
		// abort if we are at EOF.
		if l.ch == 0 {
			found = true
		}

		// keep going until we find "*/"
		if l.ch == '*' && l.peekChar() == '/' {
			found = true

			// Since the end sequence uses two characters,
			// we need to consume both.
			l.readChar()
		}
		l.readChar()
	}

	l.skipWhitespace()
}

func (l *Lexer) readItentifier() string {
	position := l.position
	
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}

	return string(l.input[position:l.position])
}

func (l *Lexer) readNumber() string {
	position := l.position
	
	for isDigit(l.ch) {
		l.readChar()
	}

	return string(l.input[position:l.position])
}

func (l *Lexer) readStringLiteral() string {
	out := ""

	for {
		l.readChar()

		if l.ch == '"' || l.ch == 0 {
			break
		}

		// Handle \n, \r, \t, \", etc.
		if l.ch == '\\' {
			l.readChar()
			if l.ch == 'n' {
				l.ch = '\n'
			}
			if l.ch == 'r' {
				l.ch = '\r'
			}
			if l.ch == 't' {
				l.ch = '\t'
			}
			if l.ch == '"' {
				l.ch = '"'
			}
			if l.ch == '\\' {
				l.ch = '\\'
			}
		}

		out = out + string(l.ch)
	}

	return out
}

func (l *Lexer) newTokenFromRune(tokenType token.TokenType, ch rune) token.Token {
	return l.newToken(tokenType, string(ch))
}

func (l *Lexer) newToken(tokenType token.TokenType, tokenLiteral string) token.Token {
	return token.Token{
		Type:     tokenType,
		Literal:  tokenLiteral,
		Position: token.TokenPosition{Line: l.line, Column: l.column},
	}
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}
