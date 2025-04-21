package token

// TokenType representa os diferentes tipos de tokens.
type TokenType string

// Token representa um único token gerado pelo lexer.
type Token struct {
	Type   TokenType
	Lexeme string
	Line   int // linha para mapear os erros
	Column int // coluna para mapear os erros
}

// Definição dos tipos de tokens.
const (
	ILLEGAL        TokenType = "ILLEGAL"    // Token inválido
	EOF            TokenType = "EOF"        // Fim do código-fonte
	IDENTIFIER     TokenType = "IDENTIFIER" // Identificadores (variáveis, funções)
	TYPE           TokenType = "TYPE"       // int, float, void
	RETURN         TokenType = "RETURN"     // return
	NUMBER         TokenType = "NUMBER"     // Números inteiros
	PLUS           TokenType = "PLUS"       // +
	MINUS          TokenType = "MINUS"      // -
	MULT           TokenType = "MULT"       // *
	DIV            TokenType = "DIV"        // /
	ASSIGN         TokenType = "ASSIGN"     // =
	SEMICOLON      TokenType = "SEMICOLON"  // ;
	LPAREN         TokenType = "LPAREN"     // (
	RPAREN         TokenType = "RPAREN"     // )
	LBRACE         TokenType = "LBRACE"     // {
	RBRACE         TokenType = "RBRACE"     // }
	IF             TokenType = "IF"         // if
	ELSE           TokenType = "ELSE"       // else
	WHILE          TokenType = "WHILE"      // while
	FOR            TokenType = "FOR"        // for
	NOT            TokenType = "NOT"        // Operador lógico NOT (!)
	NOT_EQ         TokenType = "NOT_EQ"     // !=
	GTE            TokenType = "GTE"        // >=
	LTE            TokenType = "LTE"        // <=
	GT             TokenType = "GT"         // >
	LT             TokenType = "LT"         // <
	EQ             TokenType = "EQ"         // == (igualdade)
	AND            TokenType = "AND"        // &&
	OR             TokenType = "OR"         // ||
	INT            TokenType = "INT"        // int
	FLOAT          TokenType = "FLOAT"      // float
	BOOLEAN        TokenType = "BOOLEAN"    // true/false
	STRING         TokenType = "STRING"     // "texto"
	STRING_LITERAL TokenType = "STRING_LITERAL"
	FUNC           TokenType = "FUNC"  // func declaration
	COMMA          TokenType = "COMMA" //  ,
	COLON          TokenType = "COLON" // :
)
