package parser

import (
	"fmt"
	"simple-compiler/token"
	"strconv"
)

// Parser estrutura que gerencia a análise sintática
type Parser struct {
	tokens  []token.Token
	pos     int
	current token.Token
}

// New cria um novo Parser
func New(tokens []token.Token) *Parser {
	p := &Parser{tokens: tokens, pos: 0}
	p.nextToken() // Inicializa o primeiro token
	return p
}

// AtEnd verifica se chegou ao final dos tokens
func (p *Parser) AtEnd() bool {
	return p.current.Type == token.EOF
}

// nextToken avança para o próximo token
func (p *Parser) nextToken() {
	if p.pos < len(p.tokens) {
		p.current = p.tokens[p.pos]
		p.pos++
	} else {
		p.current = token.Token{Type: token.EOF, Lexeme: ""}
	}
}

// Skip avança para evitar loops infinitos em caso de erro
func (p *Parser) Skip() {
	if !p.AtEnd() {
		p.nextToken()
	}
}

// parseExpression analisa expressões (números, identificadores, expressões binárias)
// parseExpression trata a expressão inteira, respeitando a precedência.
func (p *Parser) parseExpression() Expression {
    return p.parseComparison() // Começa com comparação, que tem a menor precedência
}

// parseComparison lida com operadores de comparação: >, <, >=, <=, ==.
func (p *Parser) parseComparison() Expression {
    left := p.parseAddition() // Inicia com a análise de adição (que tem precedência maior)

    for p.current.Type == token.GT || p.current.Type == token.LT ||
        p.current.Type == token.GTE || p.current.Type == token.LTE ||
        p.current.Type == token.EQ {

        op := p.current.Lexeme
        p.nextToken()
        right := p.parseAddition() // Operação de adição (que tem maior precedência que comparação)

        left = &BinaryExpression{Left: left, Operator: op, Right: right}
    }

    return left
}

// parseAddition lida com operadores + e -.
func (p *Parser) parseAddition() Expression {
    left := p.parseMultiplication() // Começa com multiplicação (maior precedência)

    for p.current.Type == token.PLUS || p.current.Type == token.MINUS {
        op := p.current.Lexeme
        p.nextToken()
        right := p.parseMultiplication()

        left = &BinaryExpression{Left: left, Operator: op, Right: right}
    }

    return left
}

// parseMultiplication lida com * e / (maior precedência).
func (p *Parser) parseMultiplication() Expression {
    left := p.parsePrimary() // Inicia com o valor primário (número ou identificador)

    for p.current.Type == token.MULT || p.current.Type == token.DIV {
        op := p.current.Lexeme
        p.nextToken()
        right := p.parsePrimary() // Análise do valor primário para multiplicação ou divisão

        left = &BinaryExpression{Left: left, Operator: op, Right: right}
    }

    return left
}

// parsePrimary lida com números, identificadores e expressões entre parênteses.
func (p *Parser) parsePrimary() Expression {
    switch p.current.Type {
    case token.NUMBER:
        value, err := strconv.ParseFloat(p.current.Lexeme, 64)
        if err != nil {
            fmt.Printf("Erro ao converter número: %v\n", err)
            return nil
        }
        expr := &Number{Value: value}
        p.nextToken()
        return expr

    case token.IDENTIFIER:
        expr := &Identifier{Name: p.current.Lexeme}
        p.nextToken()
        return expr

    case token.LPAREN:
        p.nextToken()
        expr := p.parseExpression() // Expressão dentro de parênteses
        if p.current.Type != token.RPAREN {
            fmt.Println("Erro: esperado ')' após expressão")
            return nil
        }
        p.nextToken()
        return expr

    default:
        fmt.Printf("Erro: token inesperado %s\n", p.current.Lexeme)
        return nil
    }
}

// ParseStatement analisa comandos como atribuições e estruturas condicionais
func (p *Parser) ParseStatement() Statement {
	switch p.current.Type {
	case token.IDENTIFIER:
		return p.ParseAssignment()

	case token.IF:
		return p.parseIfStatement()

	default:
		fmt.Printf("Erro: declaração inválida com token %s\n", p.current.Lexeme)
		return nil
	}
}

// ParseAssignment analisa expressões de atribuição (ex: x = 10)
func (p *Parser) ParseAssignment() Statement {
	name := p.current.Lexeme
	p.nextToken()

	if p.current.Type != token.ASSIGN {
		fmt.Println("Erro: esperado '=' após identificador")
		return nil
	}
	p.nextToken()

	value := p.parseExpression()
	if value == nil {
		fmt.Println("Erro: esperado valor após '='")
		return nil
	}

	return &AssignmentStatement{Name: name, Value: value}
}

// parseIfStatement analisa comandos `if`
func (p *Parser) parseIfStatement() *IfStatement {
	p.nextToken() // Consumir "if"

	condition := p.parseExpression()
	if condition == nil {
		fmt.Println("Erro: esperado condição válida após 'if'")
		return nil
	}

	// Verificar se há '{' antes do bloco
	if p.current.Type != token.LBRACE {
		fmt.Printf("Erro: esperado '{' após condição, mas encontrado '%s'\n", p.current.Lexeme)
		return nil
	}
	p.nextToken() // Consumir '{'

	var body []Statement
	for p.current.Type != token.RBRACE && p.current.Type != token.EOF {
		stmt := p.ParseStatement()
		if stmt != nil {
			body = append(body, stmt)
		} else {
			p.Skip() // Evitar erro de parsing travar tudo
		}
	}

	// Certificar que temos um '}'
	if p.current.Type != token.RBRACE {
		fmt.Println("Erro: esperado '}' ao final do bloco 'if'")
		return nil
	}
	p.nextToken() // Consumir '}'

	// Se não houver corpo dentro do if, retorna erro
	if len(body) == 0 {
		fmt.Println("Erro: bloco 'if' vazio")
		return nil
	}

	return &IfStatement{Condition: condition, Body: body}
}

// Parse inicia o processo de parsing e retorna a AST
func (p *Parser) Parse() []Statement {
	var statements []Statement

	for !p.AtEnd() {
		stmt := p.ParseStatement()
		if stmt != nil {
			statements = append(statements, stmt)
		} else {
			p.Skip() // Evita erro de parsing travar tudo
		}
	}

	return statements
}
