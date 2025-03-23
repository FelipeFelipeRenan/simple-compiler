package lexer

import (
	"simple-compiler/token"
	"unicode"
)

// Lexer representa o analisador léxico
type Lexer struct {
	input        string
	position     int  // Posição atual no input
	readPosition int  // Próxima posição a ser lida
	ch           byte // Caractere atual
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
		l.ch = 0 // Fim do input
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

// Avança e ignora espaços em branco
func (l *Lexer) skipWhitespace() {
	for unicode.IsSpace(rune(l.ch)) {
		l.readChar()
	}
}

// Lê um identificador (variável, palavra-chave)
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
		tok = token.Token{Type: token.ASSIGN, Lexeme: "="}
	case ';':
		tok = token.Token{Type: token.SEMICOLON, Lexeme: ";"}
	case '(':
		tok = token.Token{Type: token.LPAREN, Lexeme: "("}
	case ')':
		tok = token.Token{Type: token.RPAREN, Lexeme: ")"}
	case 0:
		tok = token.Token{Type: token.EOF, Lexeme: ""}
	default:
		if unicode.IsLetter(rune(l.ch)) {
			lexeme := l.readIdentifier()
			tok = token.Token{Type: token.IDENTIFIER, Lexeme: lexeme}
			return tok
		} else if unicode.IsDigit(rune(l.ch)) {
			lexeme := l.readNumber()
			tok = token.Token{Type: token.NUMBER, Lexeme: lexeme}
			return tok
		} else {
			tok = token.Token{Type: token.ILLEGAL, Lexeme: string(l.ch)}
		}
	}

	l.readChar() // AVANÇA para o próximo caractere!
	return tok
}

func Tokenize(input string)[]token.Token{
	lexer := New(input)
	tokens := []token.Token{}

	for{
		tok := lexer.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == token.EOF {
			break
		}
	}
	return tokens
}