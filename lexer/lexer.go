package lexer

import (
	"simple-compiler/token"
	"unicode"
)

type Lexer struct {
	input        string
	position     int  // posição atual a ser lida
	readPosition int  // proxima posição a ser lida (lookahead)
	ch           byte // caractere atual
}

// cria um novo lexer
func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

// função para avançar na leitura da entrada
func (l *Lexer) readChar() {
	
	if l.readPosition >= len(l.input) {
		l.ch = 0 // fim do input
	} else {
		l.ch = l.input[l.readPosition]
	}

	l.position = l.readPosition
	l.readPosition++
}

func (l *Lexer) skipWhitespaces(){
	for unicode.IsSpace(rune(l.ch)){
		l.readChar()
	}
}

// lê um identificador (palavra-chave, variavel)
func (l *Lexer) readIdentifier() string{
	start := l.position
	for unicode.IsLetter(rune(l.ch)){
		l.readChar()
	}
	return l.input[start:l.position]
}

// lê um numero
func (l *Lexer) readNumber() string{
	start := l.position
	for unicode.IsDigit(rune(l.ch)){
		l.readChar()
	}
	return l.input[start:l.position]
}

// proximo token do input
func (l *Lexer) NextToken() token.Token{
	l.skipWhitespaces()

	switch l.ch {
	case '+':
		return token.Token{Type: token.PLUS, Lexeme: "+"}
	case '-':
		return token.Token{Type: token.MINUS, Lexeme: "-"}
	case '*':
		return token.Token{Type: token.MULT, Lexeme: "*"}
	case '/':
		return token.Token{Type: token.DIV, Lexeme: "/"}
	case '=':
		return token.Token{Type: token.ASSIGN, Lexeme: "="}
	case ';':
		return token.Token{Type: token.SEMICOLON, Lexeme: ";"}
	case '(':
		return token.Token{Type: token.LPAREN, Lexeme: "("}
	case ')':
		return token.Token{Type: token.RPAREN, Lexeme: ")"}
	case 0:
		return token.Token{Type: token.EOF, Lexeme: ""}
	default:
		if unicode.IsLetter(rune(l.ch)){
			lexeme := l.readIdentifier()
			return token.Token{Type: token.IDENT, Lexeme: lexeme}
		}else if unicode.IsDigit(rune(l.ch)) {
			lexeme := l.readNumber()
			return token.Token{Type: token.NUMBER, Lexeme: lexeme}
		}else {
			return token.Token{Type: token.ILLEGAL, Lexeme: string(l.ch)}
		}

	}
}

