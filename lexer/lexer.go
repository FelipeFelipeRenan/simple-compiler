package lexer

import (
	"simple-compiler/token"
	"unicode"
)

// Lexer representa o analisador léxico
type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
}

// New cria um novo lexer
func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

// Avança para o próximo caractere
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

// Pula espaços em branco
func (l *Lexer) skipWhitespace() {
	for unicode.IsSpace(rune(l.ch)) {
		l.readChar()
	}
}

// Palavras-chave
var keywords = map[string]token.TokenType{
	"if":   token.IF,
	"else": token.ELSE,
	"for" : token.FOR,
	"while": token.WHILE,
}

// Lê um identificador e verifica se é uma palavra-chave
func (l *Lexer) readIdentifier() string {
	start := l.position
	for unicode.IsLetter(rune(l.ch)) {
		l.readChar()
	}
	return l.input[start:l.position]
}

// Lê um número
func (l *Lexer) readNumber() string {
	start := l.position
	for unicode.IsDigit(rune(l.ch)) {
		l.readChar()
	}
	return l.input[start:l.position]
}

// Lê o próximo caractere sem avançar
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// Próximo token do input
func (l *Lexer) NextToken() token.Token {
	l.skipWhitespace()

	var tok token.Token

	switch l.ch {
	case '+':
		tok = token.Token{Type: token.PLUS, Lexeme: "+"}
	case '-':
		tok = token.Token{Type: token.MINUS, Lexeme: "-"}
	case '*':
		tok = token.Token{Type: token.MULT, Lexeme: "*"}
	case '/':
		tok = token.Token{Type: token.DIV, Lexeme: "/"}
	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.EQ, Lexeme: "=="}
		} else {
			tok = token.Token{Type: token.ASSIGN, Lexeme: "="}
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.GTE, Lexeme: ">="}
		} else {
			tok = token.Token{Type: token.GT, Lexeme: ">"}
		}
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.LTE, Lexeme: "<="}
		} else {
			tok = token.Token{Type: token.LT, Lexeme: "<"}
		}
	case ';':
		tok = token.Token{Type: token.SEMICOLON, Lexeme: ";"}
	case '(':
		tok = token.Token{Type: token.LPAREN, Lexeme: "("}
	case ')':
		tok = token.Token{Type: token.RPAREN, Lexeme: ")"}
	case '{':
		tok = token.Token{Type: token.LBRACE, Lexeme: "{"}
	case '}':
		tok = token.Token{Type: token.RBRACE, Lexeme: "}"}
	case 0:
		tok = token.Token{Type: token.EOF, Lexeme: ""}
	default:
		if unicode.IsLetter(rune(l.ch)) {
			lexeme := l.readIdentifier()
			tokType := token.IDENTIFIER
			if keyword, ok := keywords[lexeme]; ok {
				tokType = keyword
			}
			tok = token.Token{Type: tokType, Lexeme: lexeme}
			return tok
		} else if unicode.IsDigit(rune(l.ch)) {
			lexeme := l.readNumber()
			tok = token.Token{Type: token.NUMBER, Lexeme: lexeme}
			return tok
		} else {
			tok = token.Token{Type: token.ILLEGAL, Lexeme: string(l.ch)}
		}
	}

	l.readChar()
	return tok
}
