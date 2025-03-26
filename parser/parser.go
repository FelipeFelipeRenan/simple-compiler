package parser

import (
	"fmt"
	"simple-compiler/token"
	"strconv"
)

// Parser estrutura que gerencia a análise sintática
type Parser struct {
	tokens      []token.Token
	pos         int
	current     token.Token
	symbolTable *SymbolTable
	errors      []string
}

func New(tokens []token.Token) *Parser {
	p := &Parser{
		tokens:      tokens,
		symbolTable: NewSymbolTable(),
		errors:      make([]string, 0),
	}
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
	case token.TYPE: // Declaração de variável
		return p.ParseVariableDeclaration()
	case token.IDENTIFIER:
		return p.ParseAssignmentOrExpression()
	case token.IF:
		return p.parseIfStatement()
	case token.WHILE:
		return p.parseWhileStatement()
	case token.FOR:
		return p.parseForStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		p.addError(fmt.Sprintf("Declaração inválida com token %s", p.current.Lexeme))
		p.Skip()
		return nil
	}
}

// ParseAssignment analisa expressões de atribuição (ex: x = 10)
// Modificação para verificar símbolos em atribuições
func (p *Parser) ParseAssignment() Statement {
	name := p.current.Lexeme

	// Verifica se a variável foi declarada
	if _, exists := p.symbolTable.Resolve(name); !exists {
		p.addError(fmt.Sprintf("Variável '%s' não declarada", name))
	}

	p.nextToken()

	if p.current.Type != token.ASSIGN {
		p.addError("Esperado '=' após identificador")
		return nil
	}
	p.nextToken()

	value := p.parseExpression()

	// Atualiza o valor na tabela de símbolos
	if err := p.symbolTable.Update(name, value); err != nil {
		p.addError(err.Error())
	}

	return &AssignmentStatement{
		Name:  name,
		Value: value,
	}
}

// parseIfStatement analisa comandos `if` com suporte a `else`
func (p *Parser) parseIfStatement() Statement {
    p.nextToken() // Pula o 'if'

    // Verifica parêntese de abertura
    if p.current.Type != token.LPAREN {
        p.addError("Esperado '(' após 'if'")
        return nil
    }
    p.nextToken()

    // Parse da condição
    condition := p.parseExpression()
    if condition == nil {
        return nil
    }

    // Verifica parêntese de fechamento
    if p.current.Type != token.RPAREN {
        p.addError("Esperado ')' após condição do if")
        return nil
    }
    p.nextToken()

    // Parse do bloco principal
    if p.current.Type != token.LBRACE {
        p.addError("Esperado '{' após condição do if")
        return nil
    }
    p.nextToken()

    var body []Statement
    for p.current.Type != token.RBRACE && !p.AtEnd() {
        stmt := p.ParseStatement()
        if stmt != nil {
            body = append(body, stmt)
        }
    }

    if p.current.Type != token.RBRACE {
        p.addError("Esperado '}' ao final do bloco if")
        return nil
    }
    p.nextToken()

    // Cria e retorna o nó IfStatement
    return &IfStatement{
        Condition: condition,
        Body:      &BlockStatement{Statements: body},
    }
}

func (p *Parser) parseWhileStatement() Statement {
    whileStmt := &WhileStatement{}
    p.nextToken() // Pula o 'while'

    // Parse condition
    if p.current.Type != token.LPAREN {
        p.addError("Esperado '(' após 'while'")
        return nil
    }
    p.nextToken()

    whileStmt.Condition = p.parseExpression()
    if whileStmt.Condition == nil {
        p.addError("Condição inválida após 'while'")
        return nil
    }

    if p.current.Type != token.RPAREN {
        p.addError("Esperado ')' após condição do 'while'")
        return nil
    }
    p.nextToken()

    // Parse body
    if p.current.Type != token.LBRACE {
        p.addError("Esperado '{' após condição do 'while'")
        return nil
    }
    p.nextToken()

    var body []Statement
    for p.current.Type != token.RBRACE && !p.AtEnd() {
        stmt := p.ParseStatement()
        if stmt != nil {
            body = append(body, stmt)
        }
    }

    if p.current.Type != token.RBRACE {
        p.addError("Esperado '}' ao final do bloco 'while'")
        return nil
    }
    p.nextToken()

    whileStmt.Body = &BlockStatement{Statements: body}
    return whileStmt
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

// Nova função para declaração de variáveis
func (p *Parser) ParseVariableDeclaration() Statement {
	typeToken := p.current
	p.nextToken()

	if p.current.Type != token.IDENTIFIER {
		p.addError("Esperado identificador após tipo")
		return nil
	}

	name := p.current.Lexeme
	p.nextToken()

	// Verifica se variável já foi declarada
	if p.symbolTable.ExistsInCurrentScope(name) {
		p.addError(fmt.Sprintf("Variável '%s' já declarada neste escopo", name))
	}

	var value Expression
	if p.current.Type == token.ASSIGN {
		p.nextToken()
		value = p.parseExpression()
	}

	// Registra na tabela de símbolos
	err := p.symbolTable.Declare(name, SymbolInfo{
		Name:     name,
		Category: Variable,
		Type:     typeToken.Lexeme,
	})

	if err != nil {
		p.addError(err.Error())
	}

	return &VariableDeclaration{
		Type:  typeToken.Lexeme,
		Name:  name,
		Value: value,
	}
}

func (p *Parser) ParseAssignmentOrExpression() Statement {
	name := p.current.Lexeme

	if p.peekToken().Type == token.ASSIGN {
		return p.ParseAssignment()
	}

	expr := p.parseExpression()

	if ident, isIdent := expr.(*Identifier); isIdent {
		if _, exists := p.symbolTable.Resolve(ident.Name); !exists {
			p.addError(fmt.Sprintf("Variável '%s' não declarada", name))
		}
	}

	return &ExpressionStatement{Expression: expr}
}

// parseReturnStatement processa declarações de retorno
func (p *Parser) parseReturnStatement() *ReturnStatement {
	stmt := &ReturnStatement{}
	p.nextToken() // Pula o 'return'

	if p.current.Type != token.SEMICOLON {
		stmt.Value = p.parseExpression()
	}

	return stmt
}

// peekToken retorna o próximo token sem consumi-lo
func (p *Parser) peekToken() token.Token {
	if p.pos < len(p.tokens)-1 {
		return p.tokens[p.pos]
	}
	return token.Token{Type: token.EOF}
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

// Função auxiliar para registrar erros
func (p *Parser) addError(msg string) {
	p.errors = append(p.errors, msg)
}
