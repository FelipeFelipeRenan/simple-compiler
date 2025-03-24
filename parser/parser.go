package parser

import (
	"fmt"
	"simple-compiler/token"
	"strconv"
)

// Estrutura que mantem os tokens e a posição atual
type Parser struct {
	tokens      []token.Token
	current     int
	symbolTable *SymbolTable
}

// New cria um novo parser
func New(tokens []token.Token) *Parser {
	return &Parser{
		tokens:      tokens,
		current:     0,
		symbolTable: NewSymbolTable()}
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
func (p *Parser) parsePrimary() Expression {
	tok := p.currentToken()

	if tok.Type == token.NUMBER {
		p.advance()

		// verificação se é um tipo inteiro ou float
		if val, err := strconv.Atoi(tok.Lexeme); err == nil {
			return &Number{Value: ValueType{Value: val, Type: TypeInt}}
		} else if val, err := strconv.ParseFloat(tok.Lexeme, 64); err == nil {
			return &Number{Value: ValueType{Value: val, Type: TypeFloat}}
		}
	}

	if tok.Type == token.IDENTIFIER {
		p.advance()
		if val, exists := p.symbolTable.Get(tok.Lexeme); exists {
			return &Identifier{Name: tok.Lexeme, Value: val}
		} else {
			fmt.Printf("Erro: variavel %s nao foi declarada\n", tok.Lexeme)
			return nil
		}
	}

	fmt.Printf("Erro: token inesperado: %s", tok.Lexeme)
	return nil
}

// função para analisar expressoes matematicas
func (p *Parser) parseExpression() Expression {
	left := p.parsePrimary()
	for p.match(token.PLUS) || p.match(token.MINUS) || p.match(token.MULT) || p.match(token.DIV) {
		operator := p.tokens[p.current-1].Lexeme // Operador
		
		right := p.parsePrimary() 

		// Garantindo que os tipos sao compativeis
		leftType := left.(*Number).Value.Type
		rightType := right.(*Number).Value.Type

		if leftType != rightType {
			fmt.Println("Erro: operação com tipos incompativeis")
			return nil
		}
		
		// segundo operando
		left = &BinaryExpression{
			Left:     left,
			Operator: operator,
			Right:    right,
		}
	}
	return left
}

// função para analise de atribuições e adiciona as variaveies na tabela de simbolos
func (p *Parser) ParseAssignment() Statement {
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

	p.symbolTable.Set(varName, value)

	fmt.Printf("Atribuição: %s = %v\n", varName, value)
	return &Assignment{
		Variable: &Identifier{Name: varName},
		Value:    value,
	}
}
