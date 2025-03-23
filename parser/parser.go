package parser

import (
	"fmt"
	"simple-compiler/ast"
	"simple-compiler/token"
	"strconv"
)

// Estrutura que mantem os tokens e a posição atual
type Parser struct {
	tokens  []token.Token
	current int
}

// New cria um novo parser
func New(tokens []token.Token) *Parser {
	return &Parser{tokens: tokens, current: 0}
}

// currentToken retorna o token atual sem avançar
func (p *Parser) currentToken() token.Token {
	if p.current < len(p.tokens) {
		return p.tokens[p.current]
	}
	return token.Token{Type: token.EOF, Lexeme: ""}
}

// Advance move para o proximo token
func (p *Parser) advance() {
	if p.current < len(p.tokens) {
		p.current++
	}
}

func (p *Parser) match(tokenType token.TokenType) bool {
	if p.currentToken().Type == tokenType {
		p.advance()
		return true
	}
	return false
}

// Função para analisar numeros e identificadores
// parsePrimary analisa números e identificadores
func (p *Parser) parsePrimary() ast.Expression {
	tok := p.currentToken()

	if tok.Type == token.NUMBER {
		p.advance()

		value, err := strconv.Atoi(tok.Lexeme)
		if err != nil {
			fmt.Printf("Erro ao converter número: %s\n", tok.Lexeme)
			return nil
		}
		return &ast.Number{Value: value}
	}

	if tok.Type == token.IDENTIFIER {
		p.advance()
		return &ast.Identifier{Name: tok.Lexeme}
	}

	fmt.Printf("Erro: token inesperado %s\n", tok.Lexeme)
	return nil
}

// função para analisar expressoes matematicas
func (p *Parser) parseExpression() ast.Expression {
	left := p.parsePrimary()
	for p.match(token.PLUS) || p.match(token.MINUS) || p.match(token.MULT) || p.match(token.DIV) {
		operator := p.tokens[p.current-1].Lexeme // Operador
		right := p.parsePrimary()                // segundo operando
		left = &ast.BinaryExpression{
			Left:     left,
			Operator: operator,
			Right:    right,
		}
	}
	return left
}

// função para analise de atribuições
func (p *Parser) ParseAssignment() ast.Statement {
	if !p.match(token.IDENTIFIER) {
		fmt.Println("Erro: Esperando um identificador")
		return nil
	}

	varName := p.tokens[p.current-1].Lexeme

	if !p.match(token.ASSIGN) {
		fmt.Println("Erro: Esperando '='")
		return nil

	}

	value := p.parseExpression()

	return &ast.Assignment{
		Variable: &ast.Identifier{Name: varName},
		Value:    value,
	}
}
