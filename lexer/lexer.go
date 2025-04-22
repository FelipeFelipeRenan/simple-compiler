package lexer

import (
	"simple-compiler/token"
)

type Lexer struct {
	input        string
	position     int  // posição atual no input (aponta para o char atual)
	readPosition int  // posição de leitura atual no input (após o char atual)
	ch           byte // char atual sendo analisado
	line         int  // número da linha atual
	column       int  // coluna atual na linha
}

func New(input string) *Lexer {
	l := &Lexer{input: input, line: 1, column: 1}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // EOF
	} else {
		l.ch = l.input[l.readPosition]
	}
	
	// Atualiza a posição e a coluna
	l.position = l.readPosition
	l.readPosition++
	
	if l.ch == '\n' {
		l.line++
		l.column = 1
	} else if l.ch != 0 {
		l.column++
	}
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

var keywords = map[string]token.TokenType{
	"func":   token.FUNC,
	"if":     token.IF,
	"else":   token.ELSE,
	"return": token.RETURN,
	"true":   token.BOOLEAN,
	"false":  token.BOOLEAN,
	"int":    token.TYPE,
	"float":  token.TYPE,
	"void":   token.TYPE,
	"bool":   token.TYPE,
	"string": token.TYPE,
	"for":    token.FOR,
	"while":  token.WHILE,
	"print":  token.IDENTIFIER,
}

func (l *Lexer) lookupIdent(ident string) token.TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return token.IDENTIFIER
}

func (l *Lexer) readIdentifier() string {
	start := l.position
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	return l.input[start:l.position]
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func (l *Lexer) readNumber() string {
	start := l.position
	for isDigit(l.ch) {
		l.readChar()
	}

	if l.ch == '.' {
		l.readChar()
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	return l.input[start:l.position]
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) readString() string {
	l.readChar() // Pula a aspa inicial
	start := l.position

	for l.ch != '"' && l.ch != 0 {
		if l.ch == '\\' { // Trata caracteres de escape
			l.readChar()
		}
		l.readChar()
	}

	if l.ch == '"' {
		str := l.input[start:l.position]
		l.readChar() // Pula a aspa final
		return str
	}
	return "" // String não finalizada
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	tok.Line = l.line
	tok.Column = l.column

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			tok.Type = token.EQ
			tok.Lexeme = "=="
			l.readChar()
		} else {
			tok.Type = token.ASSIGN
			tok.Lexeme = "="
		}
	case '!':
		if l.peekChar() == '=' {
			tok.Type = token.NOT_EQ
			tok.Lexeme = "!="
			l.readChar()
		} else {
			tok.Type = token.NOT
			tok.Lexeme = "!"
		}
	case '<':
		if l.peekChar() == '=' {
			tok.Type = token.LTE
			tok.Lexeme = "<="
			l.readChar()
		} else {
			tok.Type = token.LT
			tok.Lexeme = "<"
		}
	case '>':
		if l.peekChar() == '=' {
			tok.Type = token.GTE
			tok.Lexeme = ">="
			l.readChar()
		} else {
			tok.Type = token.GT
			tok.Lexeme = ">"
		}
	case '+':
		tok.Type = token.PLUS
		tok.Lexeme = "+"
	case '-':
		tok.Type = token.MINUS
		tok.Lexeme = "-"
	case '*':
		tok.Type = token.MULT
		tok.Lexeme = "*"
	case '/':
		tok.Type = token.DIV
		tok.Lexeme = "/"
	case ',':
		tok.Type = token.COMMA
		tok.Lexeme = ","
	case ';':
		tok.Type = token.SEMICOLON
		tok.Lexeme = ";"
	case '(':
		tok.Type = token.LPAREN
		tok.Lexeme = "("
	case ')':
		tok.Type = token.RPAREN
		tok.Lexeme = ")"
	case '{':
		tok.Type = token.LBRACE
		tok.Lexeme = "{"
	case '}':
		tok.Type = token.RBRACE
		tok.Lexeme = "}"
	case '"':
		tok.Type = token.STRING_LITERAL
		tok.Lexeme = l.readString()
		if tok.Lexeme == "" {
			tok.Type = token.ILLEGAL
			tok.Lexeme = "string não finalizada"
		}
		return tok
	case 0:
		tok.Type = token.EOF
		tok.Lexeme = ""
	default:
		if isLetter(l.ch) {
			tok.Lexeme = l.readIdentifier()
			tok.Type = l.lookupIdent(tok.Lexeme)
			return tok
		} else if isDigit(l.ch) {
			tok.Type = token.NUMBER
			tok.Lexeme = l.readNumber()
			return tok
		} else {
			tok.Type = token.ILLEGAL
			tok.Lexeme = string(l.ch)
		}
	}

	l.readChar()
	return tok
}