package lexer

import (
	"simple-compiler/token"
	"unicode"
)

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
	line         int
	column       int
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
	l.position = l.readPosition
	l.readPosition++

	// Atualiza posição apenas se não for EOF
	if l.ch != 0 {
		if l.ch == '\n' {
			l.line++
			l.column = 1
		} else {
			l.column++
		}
	}
}

func (l *Lexer) skipWhitespace() {
	for unicode.IsSpace(rune(l.ch)) && l.ch != 0 {
		// Mantém o controle preciso de linhas/colunas
		if l.ch == '\n' {
			l.line++
			l.column = 0
		}
		l.readChar()
	}
}

var keywords = map[string]token.TokenType{
	"if":     token.IF,
	"else":   token.ELSE,
	"for":    token.FOR,
	"while":  token.WHILE,
	"int":    token.TYPE,
	"float":  token.TYPE,
	"void":   token.TYPE,
	"return": token.RETURN,
	"string": token.TYPE,
	"bool":   token.TYPE,
	"true":   token.BOOLEAN,
	"false":  token.BOOLEAN,
}

func (l *Lexer) readIdentifier() string {
	start := l.position
	for unicode.IsLetter(rune(l.ch)) || unicode.IsDigit(rune(l.ch)) || l.ch == '_' {
		l.readChar()
	}
	return l.input[start:l.position]
}

func (l *Lexer) readString() string {
	l.readChar() // Pula a aspa inicial
	start := l.position

	for l.ch != '"' && l.ch != 0 {
		l.readChar()
	}

	if l.ch == '"' {
		str := l.input[start:l.position]
		l.readChar() // Pula a aspa final
		if str == "" {
			return " " // Retorna espaço para strings vazias
		}
		return str
	}
	return "" // Indica string não finalizada
}

func (l *Lexer) readNumber() string {
	start := l.position
	for unicode.IsDigit(rune(l.ch)) {
		l.readChar()
	}

	// Parte decimal
	if l.ch == '.' {
		l.readChar()
		for unicode.IsDigit(rune(l.ch)) {
			l.readChar()
		}
	}

	return l.input[start:l.position]
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) NextToken() token.Token {
	l.skipWhitespace()

	tok := token.Token{
		Line:   l.line,
		Column: l.column,
	}
	switch l.ch {
	case 0:
		tok.Type = token.EOF
		tok.Lexeme = ""
    case '"':
        str := l.readString()
        if str == "" {
            tok.Type = token.ILLEGAL
            tok.Lexeme = "string não finalizada"
        } else {
            tok.Type = token.STRING_LITERAL
            tok.Lexeme = str
        }
        return tok
	case '=':
		if l.peekChar() == '=' {
			tok.Lexeme = "=="
			tok.Type = token.EQ
			l.readChar()
		} else {
			tok.Lexeme = "="
			tok.Type = token.ASSIGN
		}
	case '!':
		if l.peekChar() == '=' {
			tok.Lexeme = "!="
			tok.Type = token.NOT_EQ
			l.readChar()
		} else {
			tok.Lexeme = "!"
			tok.Type = token.NOT
		}
	case '+':
		tok.Lexeme = "+"
		tok.Type = token.PLUS
	case '-':
		tok.Lexeme = "-"
		tok.Type = token.MINUS
	case '*':
		tok.Lexeme = "*"
		tok.Type = token.MULT
	case '/':
		tok.Lexeme = "/"
		tok.Type = token.DIV
	case '>':
		if l.peekChar() == '=' {
			tok.Lexeme = ">="
			tok.Type = token.GTE
			l.readChar()
		} else {
			tok.Lexeme = ">"
			tok.Type = token.GT
		}
	case '<':
		if l.peekChar() == '=' {
			tok.Lexeme = "<="
			tok.Type = token.LTE
			l.readChar()
		} else {
			tok.Lexeme = "<"
			tok.Type = token.LT
		}
	case ';':
		tok.Lexeme = ";"
		tok.Type = token.SEMICOLON
	case '(':
		tok.Lexeme = "("
		tok.Type = token.LPAREN
	case ')':
		tok.Lexeme = ")"
		tok.Type = token.RPAREN
	case '{':
		tok.Lexeme = "{"
		tok.Type = token.LBRACE
	case '}':
		tok.Lexeme = "}"
		tok.Type = token.RBRACE
    default:
        if unicode.IsLetter(rune(l.ch)) {
            tok.Lexeme = l.readIdentifier()
            if kw, ok := keywords[tok.Lexeme]; ok {
                tok.Type = kw
            } else {
                tok.Type = token.IDENTIFIER
            }
            return tok
        } else if unicode.IsDigit(rune(l.ch)) {
            tok.Lexeme = l.readNumber()
            tok.Type = token.NUMBER
            return tok
        } else if l.ch != 0 {
            tok.Type = token.ILLEGAL
            tok.Lexeme = string(l.ch)
            l.readChar()
        }
    }

    if tok.Type == "" {
        tok.Type = token.EOF
    }
    
    if tok.Type != token.ILLEGAL && tok.Type != token.EOF {
        l.readChar()
    }
    
    return tok
}