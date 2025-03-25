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

// parseExpression trata a expressão inteira, respeitando a precedência.
func (p *Parser) parseExpression() Expression {
	return p.parseComparison() // Começa com comparação, que tem a menor precedência
}

// parseComparison lida com operadores de comparação: >, <, >=, <=, ==.
func (p *Parser) parseComparison() Expression {
	left := p.parseAddition()

	for p.current.Type == token.GT || p.current.Type == token.LT ||
		p.current.Type == token.GTE || p.current.Type == token.LTE ||
		p.current.Type == token.EQ {

		op := p.current.Lexeme
		p.nextToken()
		right := p.parseAddition()

		left = &BinaryExpression{Left: left, Operator: op, Right: right}
	}

	return left
}

// parseAddition lida com operadores + e -.
func (p *Parser) parseAddition() Expression {
	left := p.parseMultiplication()

	for p.current.Type == token.PLUS || p.current.Type == token.MINUS {
		op := p.current.Lexeme
		p.nextToken()
		right := p.parseMultiplication()

		left = &BinaryExpression{Left: left, Operator: op, Right: right}
	}

	return left
}

// parseMultiplication lida com * e /.
func (p *Parser) parseMultiplication() Expression {
	left := p.parsePrimary()

	for p.current.Type == token.MULT || p.current.Type == token.DIV {
		op := p.current.Lexeme
		p.nextToken()
		right := p.parsePrimary()

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
		expr := p.parseExpression()
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

	case token.WHILE:
		return p.parseWhileStatement()

	case token.FOR:
		return p.parseForStatement()

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

// parseIfStatement analisa comandos `if` com suporte a `else`
func (p *Parser) parseIfStatement() *IfStatement {
	p.nextToken()

	condition := p.parseExpression()
	if condition == nil {
		fmt.Println("Erro: esperado condição válida após 'if'")
		return nil
	}

	if p.current.Type != token.LBRACE {
		fmt.Println("Erro: esperado '{' após condição")
		return nil
	}
	p.nextToken()

	var body []Statement
	for p.current.Type != token.RBRACE && p.current.Type != token.EOF {
		stmt := p.ParseStatement()
		if stmt != nil {
			body = append(body, stmt)
		} else {
			p.Skip()
		}
	}

	if p.current.Type != token.RBRACE {
		fmt.Println("Erro: esperado '}' ao final do bloco 'if'")
		return nil
	}
	p.nextToken()

	var elseBody []Statement
	if p.current.Type == token.ELSE {
		p.nextToken()

		if p.current.Type != token.LBRACE {
			fmt.Println("Erro: esperado '{' após 'else'")
			return nil
		}
		p.nextToken()

		for p.current.Type != token.RBRACE && p.current.Type != token.EOF {
			stmt := p.ParseStatement()
			if stmt != nil {
				elseBody = append(elseBody, stmt)
			} else {
				p.Skip()
			}
		}

		if p.current.Type != token.RBRACE {
			fmt.Println("Erro: esperado '}' ao final do bloco 'else'")
			return nil
		}
		p.nextToken()
	}

	return &IfStatement{Condition: condition, Body: body, ElseBody: elseBody}
}

// parseWhileStatement analisa o comando `while`
// Espera-se que a estrutura seja: while (condição) { corpo }
func (p *Parser) parseWhileStatement() *WhileStatement {
	// Pula a palavra-chave "while"
	p.nextToken()

	// Lê a condição entre parênteses
	if p.current.Type != token.LPAREN {
		fmt.Println("Erro: esperado '(' após 'while'")
		return nil
	}
	p.nextToken()

	// Condição da expressão
	condition := p.parseExpression()
	if condition == nil {
		fmt.Println("Erro: esperado condição válida após '('")
		return nil
	}

	// Espera o fechamento do parêntese
	if p.current.Type != token.RPAREN {
		fmt.Println("Erro: esperado ')' após condição")
		return nil
	}
	p.nextToken()

	// Espera pelo bloco de código
	if p.current.Type != token.LBRACE {
		fmt.Println("Erro: esperado '{' após condição do 'while'")
		return nil
	}
	p.nextToken()

	// Lê o corpo do while
	var body []Statement
	for p.current.Type != token.RBRACE && p.current.Type != token.EOF {
		stmt := p.ParseStatement()
		if stmt != nil {
			body = append(body, stmt)
		} else {
			p.Skip()
		}
	}

	// Espera o fechamento do bloco
	if p.current.Type != token.RBRACE {
		fmt.Println("Erro: esperado '}' ao final do bloco 'while'")
		return nil
	}
	p.nextToken()

	// Retorna a estrutura do while
	return &WhileStatement{
		Condition: condition,
		Body:      body,
	}
}
// parseForStatement analisa o `for`
func (p *Parser) parseForStatement() *ForStatement {
    p.nextToken()

    if p.current.Type != token.LPAREN {
        fmt.Println("Erro: esperado '(' após 'for'")
        return nil
    }
    p.nextToken()

    // Inicialização (como atribuição)
    init := p.ParseAssignment()
    
    if p.current.Type != token.SEMICOLON {
        fmt.Println("Erro: esperado ';' após a inicialização do 'for'")
        return nil
    }
    p.nextToken()

    // Condição (expressão booleana)
    condition := p.parseExpression()

    if p.current.Type != token.SEMICOLON {
        fmt.Println("Erro: esperado ';' após a condição do 'for'")
        return nil
    }
    p.nextToken()

    // Atualização (como atribuição)
    update := p.ParseAssignment()

    if p.current.Type != token.RPAREN {
        fmt.Println("Erro: esperado ')' após cabeçalho do 'for'")
        return nil
    }
    p.nextToken()

    if p.current.Type != token.LBRACE {
        fmt.Println("Erro: esperado '{' após cabeçalho do 'for'")
        return nil
    }
    p.nextToken()

    // Corpo do 'for'
    var body []Statement
    for p.current.Type != token.RBRACE && p.current.Type != token.EOF {
        body = append(body, p.ParseStatement())
    }
    p.nextToken()

    return &ForStatement{
        Init:      init,
        Condition: condition,
        Update:    update,
        Body:      body,
    }
}


// Parse analisa uma sequência de declarações e retorna a AST completa
func (p *Parser) Parse() []Statement {
	var statements []Statement

	for !p.AtEnd() {
		stmt := p.ParseStatement()
		if stmt != nil {
			statements = append(statements, stmt)
		} else {
			p.Skip() // Evita travamento ao encontrar erro
		}
	}

	return statements
}
