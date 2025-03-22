package token

// TokenType representa os diferentes tipos de token
type TokenType string

// Token representa um único token gerado pelo lexer
type Token struct {
	Type TokenType
	Lexeme string
}


// tipos de tokens
const (
	ILLEGAL   TokenType = "ILLEGAL"   // Token inválido
	EOF       TokenType = "EOF"       // Fim do código-fonte
	IDENT     TokenType = "IDENT"     // Identificadores (variáveis, funções)
	NUMBER    TokenType = "NUMBER"    // Números inteiros
	PLUS      TokenType = "PLUS"      // +
	MINUS     TokenType = "MINUS"     // -
	MULT      TokenType = "MULT"      // *
	DIV       TokenType = "DIV"       // /
	ASSIGN    TokenType = "ASSIGN"    // =
	SEMICOLON TokenType = "SEMICOLON" // ;
	LPAREN    TokenType = "LPAREN"    // (
	RPAREN    TokenType = "RPAREN"    // )
)