package parser

import (
	"fmt"
	"simple-compiler/token"
	"strconv"
)

// Estrutura que mantém os tokens e a posição atual
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
		symbolTable: NewSymbolTable(),
	}
}

// currentToken retorna o token atual sem avançar
func (p *Parser) currentToken() token.Token {
	if p.current < len(p.tokens) {
		return p.tokens[p.current]
	}
	return token.Token{Type: token.EOF, Lexeme: ""}
}

// Advance move para o próximo token
func (p *Parser) advance() {
	if p.current < len(p.tokens) {
		p.current++
	}
}

// match verifica se o token atual é do tipo esperado e avança
func (p *Parser) match(tokenType token.TokenType) bool {
	if p.currentToken().Type == tokenType {
		p.advance()
		return true
	}
	return false
}

// parsePrimary analisa números e identificadores
func (p *Parser) parsePrimary() Expression {
	tok := p.currentToken()

	if tok.Type == token.NUMBER {
		p.advance()
		val, err := strconv.ParseFloat(tok.Lexeme, 64)
		if err == nil {
			return &Number{Value: ValueType{Value: val, Type: TypeFloat}}
		}
		fmt.Printf("Erro: número inválido %s\n", tok.Lexeme)
		return nil
	}

	if tok.Type == token.IDENTIFIER {
		p.advance()
		if val, exists := p.symbolTable.Get(tok.Lexeme); exists {
			if expr, ok := val.(Expression); ok {
				return expr
			}
			fmt.Printf("Erro: variável %s contém um tipo inválido\n", tok.Lexeme)
			return nil
		}
		fmt.Printf("Erro: variável %s não foi declarada\n", tok.Lexeme)
		return nil
	}

	fmt.Printf("Erro: token inesperado: %s\n", tok.Lexeme)
	return nil
}

// parseFactor analisa multiplicação e divisão com precedência
func (p *Parser) parseFactor() Expression {
	left := p.parsePrimary()
	if left == nil {
		return nil
	}

	for p.match(token.MULT) || p.match(token.DIV) {
		operator := p.tokens[p.current-1].Lexeme
		right := p.parsePrimary()
		if right == nil {
			return nil
		}
		left = &BinaryExpression{Left: left, Operator: operator, Right: right}
	}

	return left
}

// parseExpression analisa adição e subtração respeitando a precedência
func (p *Parser) parseExpression() Expression {
	left := p.parseFactor()
	if left == nil {
		return nil
	}

	for p.match(token.PLUS) || p.match(token.MINUS) {
		operator := p.tokens[p.current-1].Lexeme
		right := p.parseFactor()
		if right == nil {
			return nil
		}
		left = &BinaryExpression{Left: left, Operator: operator, Right: right}
	}

	return left
}

// parseAssignment analisa atribuições e adiciona variáveis à tabela de símbolos
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
	if value == nil {
		fmt.Println("Erro: expressão inválida")
		return nil
	}

	if num, ok := value.(*Number); ok {
		p.symbolTable.Set(varName, num)
	} else if expr, ok := value.(*BinaryExpression); ok {
		evaluated := evaluateExpression(expr)
		if evaluated != nil {
			p.symbolTable.Set(varName, evaluated)
		} else {
			fmt.Println("Erro: falha ao avaliar expressão matemática")
			return nil
		}
	} else {
		fmt.Println("Erro: Apenas números e expressões matemáticas são suportados na atribuição")
		return nil
	}

	//	fmt.Printf("Atribuição: %s = %v\n", varName, value)
	return &Assignment{
		Variable: &Identifier{Name: varName},
		Value:    value,
	}
}

// evaluateExpression avalia expressões matemáticas recursivamente
func evaluateExpression(expr Expression) *Number {
	switch e := expr.(type) {
	case *Number:
		return e
	case *BinaryExpression:
		left := evaluateExpression(e.Left)
		right := evaluateExpression(e.Right)

		if left == nil || right == nil {
			fmt.Println("Erro: Apenas números são suportados em expressões matemáticas")
			return nil
		}

		leftVal := left.Value.Value
		rightVal := right.Value.Value

		var result float64
		switch e.Operator {
		case "+":
			result = leftVal + rightVal
		case "-":
			result = leftVal - rightVal
		case "*":
			result = leftVal * rightVal
		case "/":
			if rightVal == 0 {
				fmt.Println("Erro: divisão por zero")
				return nil
			}
			result = leftVal / rightVal
		default:
			fmt.Println("Erro: Operador desconhecido")
			return nil
		}

		return &Number{Value: ValueType{Value: result, Type: TypeFloat}}
	default:
		fmt.Println("Erro: Expressão inválida")
		return nil
	}
}
