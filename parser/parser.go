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
	Errors      []ParseError
}

type ParseError struct {
	Message string
	Column  int
	Line    int
	Token   string
}

func New(tokens []token.Token) *Parser {
    p := &Parser{
        tokens:      tokens,
        symbolTable: NewSymbolTable(), // Certifique-se que está inicializando aqui
        Errors:      make([]ParseError, 0),
    }
    p.nextToken()
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
    return p.parseAssignment()
}

func (p *Parser) parseAssignment() Expression {
    expr := p.parseLogicalOr()
    
    if p.current.Type == token.ASSIGN {
        if ident, ok := expr.(*Identifier); ok {
            opToken := p.current
            p.nextToken()
            value := p.parseAssignment()
            return &BinaryExpression{
                Left:     ident,
                Operator: opToken.Lexeme,
                Right:    value,
                Token:    opToken,
            }
        }
        p.addError("Esperado identificador no lado esquerdo da atribuição", 
                  p.current.Line, p.current.Column)
    }
    
    return expr
}

func (p *Parser) ParseAssignment() Statement {
    name := p.current.Lexeme
    currentToken := p.current

    // Verifica se a variável foi declarada
    if _, exists := p.symbolTable.Resolve(name); !exists {
        p.addError(fmt.Sprintf("Variável '%s' não declarada", name), 
                  currentToken.Line, currentToken.Column)
    }

    p.nextToken() // Pula o nome da variável

    // CORREÇÃO: Compare p.current.Type com token.ASSIGN
    if p.current.Type != token.ASSIGN {
        p.addError("Esperado '=' após identificador", currentToken.Line, currentToken.Column)
        return nil
    }
    
    assignToken := p.current
    p.nextToken() // Pula o '='

    value := p.parseExpression()

    return &AssignmentStatement{
        Name:  name,
        Value: value,
        Token: assignToken,
    }
}

func (p *Parser) ParseAssignmentOrExpression() Statement {
    name := p.current.Lexeme

    // Verifica se é uma atribuição
    if p.peekToken().Type == token.ASSIGN {
        return p.ParseAssignment()
    }

    // Verifica se a variável foi declarada
    if _, exists := p.symbolTable.Resolve(name); !exists {
        p.addError(fmt.Sprintf("Variável '%s' não declarada", name), 
                  p.current.Line, p.current.Column)
    }

    // Processa como expressão
    expr := p.parseExpression()
    return &ExpressionStatement{Expression: expr}
}

func (p *Parser) parseLogicalOr() Expression {
    expr := p.parseLogicalAnd()
    
    for p.current.Type == token.OR {
        opToken := p.current
        p.nextToken()
        right := p.parseLogicalAnd()
        expr = &BinaryExpression{
            Left:     expr,
            Operator: opToken.Lexeme,
            Right:    right,
            Token:    opToken,
        }
    }
    
    return expr
}

func (p *Parser) parseLogicalAnd() Expression {
    expr := p.parseEquality()
    
    for p.current.Type == token.AND {
        opToken := p.current
        p.nextToken()
        right := p.parseEquality()
        expr = &BinaryExpression{
            Left:     expr,
            Operator: opToken.Lexeme,
            Right:    right,
            Token:    opToken,
        }
    }
    
    return expr
}

func (p *Parser) parseEquality() Expression {
    expr := p.parseComparison()
    
    for p.current.Type == token.EQ {
        opToken := p.current
        p.nextToken()
        right := p.parseComparison()
        expr = &BinaryExpression{
            Left:     expr,
            Operator: opToken.Lexeme,
            Right:    right,
            Token:    opToken,
        }
    }
    
    return expr
}

func (p *Parser) parseComparison() Expression {
    expr := p.parseAddition()
    
    for p.current.Type == token.LT || p.current.Type == token.LTE || 
        p.current.Type == token.GT || p.current.Type == token.GTE {
        opToken := p.current
        p.nextToken()
        right := p.parseAddition()
        expr = &BinaryExpression{
            Left:     expr,
            Operator: opToken.Lexeme,
            Right:    right,
            Token:    opToken,
        }
    }
    
    return expr
}

func (p *Parser) parseAddition() Expression {
    expr := p.parseMultiplication()
    
    for p.current.Type == token.PLUS || p.current.Type == token.MINUS {
        opToken := p.current
        p.nextToken()
        right := p.parseMultiplication()
        expr = &BinaryExpression{
            Left:     expr,
            Operator: opToken.Lexeme,
            Right:    right,
            Token:    opToken,
        }
    }
    
    return expr
}

func (p *Parser) parseMultiplication() Expression {
    expr := p.parseUnary()
    
    for p.current.Type == token.MULT || p.current.Type == token.DIV {
        opToken := p.current
        p.nextToken()
        right := p.parseUnary()
        expr = &BinaryExpression{
            Left:     expr,
            Operator: opToken.Lexeme,
            Right:    right,
            Token:    opToken,
        }
    }
    
    return expr
}

func (p *Parser) parseUnary() Expression {
    if p.current.Type == token.MINUS || p.current.Type == token.NOT {
        opToken := p.current
        p.nextToken()
        return &UnaryExpression{
            Operator: opToken.Lexeme,
            Right:    p.parsePrimary(),
            Token:    opToken,
        }
    }
    return p.parsePrimary()
}

func (p *Parser) parsePrimary() Expression {
    switch p.current.Type {
    case token.NUMBER:
        value, err := strconv.ParseFloat(p.current.Lexeme, 64)
        if err != nil {
            p.addError(fmt.Sprintf("Erro ao converter número: %v", err),
                      p.current.Line, p.current.Column)
            return nil
        }
        expr := &Number{Value: value, Token: p.current}
        p.nextToken()
        return expr
        
    case token.IDENTIFIER:
        expr := &Identifier{Name: p.current.Lexeme, Token: p.current}
        p.nextToken()
        return expr
        
    case token.LPAREN:
        p.nextToken()
        expr := p.parseExpression()
        if p.current.Type != token.RPAREN {
            p.addError("Esperado ')' após expressão", p.current.Line, p.current.Column)
            return nil
        }
        p.nextToken()
        return expr
        
    case token.BOOLEAN:
        value := p.current.Lexeme == "true"
        expr := &BooleanLiteral{Value: value, Token: p.current}
        p.nextToken()
        return expr
        
    case token.STRING_LITERAL:
        expr := &StringLiteral{Value: p.current.Lexeme, Token: p.current}
        p.nextToken()
        return expr
        
    default:
        p.addError(fmt.Sprintf("Token inesperado: %s", p.current.Lexeme),
                  p.current.Line, p.current.Column)
        p.nextToken()
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
    case token.FUNC:
        return p.parseFunctionDeclaration()
	default:
		p.addError(fmt.Sprintf("Declaração inválida com token %s", p.current.Lexeme), p.current.Line, p.current.Column)
		p.Skip()
		return nil
	}
}

// ParseAssignment analisa expressões de atribuição (ex: x = 10)

func (p *Parser) parseIfStatement() Statement {
	p.nextToken() // Pula o 'if'

	// Verifica parêntese de abertura
	if p.current.Type != token.LPAREN {
		p.addError("Esperado '(' após 'if'", p.current.Line, p.current.Column)
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
		p.addError("Esperado ')' após condição do if", p.current.Line, p.current.Column)
		return nil
	}
	p.nextToken()

	// Parse do bloco principal
	if p.current.Type != token.LBRACE {
		p.addError("Esperado '{' após condição do if", p.current.Line, p.current.Column)
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
		p.addError("Esperado '}' ao final do bloco if", p.current.Line, p.current.Column)
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
		p.addError("Esperado '(' após 'while'", p.current.Line, p.current.Column)
		return nil
	}
	p.nextToken()

	whileStmt.Condition = p.parseExpression()
	if whileStmt.Condition == nil {
		p.addError("Condição inválida após 'while'", p.current.Line, p.current.Column)
		return nil
	}

	if p.current.Type != token.RPAREN {
		p.addError("Esperado ')' após condição do 'while'", p.current.Line, p.current.Column)
		return nil
	}
	p.nextToken()

	// Parse body
	if p.current.Type != token.LBRACE {
		p.addError("Esperado '{' após condição do 'while'", p.current.Line, p.current.Column)
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
		p.addError("Esperado '}' ao final do bloco 'while'", p.current.Line, p.current.Column)
		return nil
	}
	p.nextToken()

	whileStmt.Body = &BlockStatement{Statements: body}
	return whileStmt
}
func (p *Parser) parseBlock() *BlockStatement {
    if p.current.Type != token.LBRACE {
        p.addError("Expected '{' to start block", p.current.Line, p.current.Column)
        return nil
    }

    block := &BlockStatement{}
    p.nextToken() // Pula '{'

    for p.current.Type != token.RBRACE && !p.AtEnd() {
        stmt := p.ParseStatement()
        if stmt != nil {
            block.Statements = append(block.Statements, stmt)
        } else {
            // Recuperação de erro
            p.skipUntil(token.SEMICOLON, token.RBRACE)
            if p.current.Type == token.SEMICOLON {
                p.nextToken()
            }
        }
    }

    if p.current.Type != token.RBRACE {
        p.addError("Expected '}' to close block", p.current.Line, p.current.Column)
    } else {
        p.nextToken() // Pula '}'
    }

    return block
}
// parseForStatement analisa o `for`
func (p *Parser) parseForStatement() Statement {
	forStmt := &ForStatement{}
	p.nextToken() // Pula 'for'

	// Exige parênteses de abertura
	if p.current.Type != token.LPAREN {
		p.addError("Esperado '(' após 'for'", p.current.Line, p.current.Column)
		return nil
	}
	p.nextToken()

	// Inicialização (pode ser vazia)
	if p.current.Type != token.SEMICOLON {
		forStmt.Init = p.ParseStatement()
	}

	// Exige primeiro ;
	if p.current.Type != token.SEMICOLON {
		p.addError("Esperado ';' após inicialização do for", p.current.Line, p.current.Column)
		return nil
	}
	p.nextToken()

	// Condição (pode ser vazia)
	if p.current.Type != token.SEMICOLON {
		forStmt.Condition = p.parseExpression()
	}

	// Exige segundo ;
	if p.current.Type != token.SEMICOLON {
		p.addError("Esperado ';' após condição do for", p.current.Line, p.current.Column)
		return nil
	}
	p.nextToken()

	// Atualização (pode ser vazia)
	if p.current.Type != token.RPAREN {
		forStmt.Update = p.ParseStatement()
	}

	// Exige parênteses de fechamento
	if p.current.Type != token.RPAREN {
		p.addError("Esperado ')' após cabeçalho do for", p.current.Line, p.current.Column)
		return nil
	}
	p.nextToken()

	// Exige bloco com {}
	if p.current.Type != token.LBRACE {
		p.addError("Esperado '{' após cabeçalho do for", p.current.Line, p.current.Column)
		return nil
	}
	forStmt.Body = p.parseBlock()

	return forStmt
}

// Nova função para declaração de variáveis
func (p *Parser) ParseVariableDeclaration() Statement {
    typeToken := p.current
    p.nextToken()

    if p.current.Type != token.IDENTIFIER {
        p.addError(fmt.Sprintf("Esperado nome da variável após tipo '%s'", typeToken.Lexeme),
                  typeToken.Line, typeToken.Column)
        return nil
    }

    nameToken := p.current
    
    // Avança para o próximo token antes de verificar a atribuição
    p.nextToken()

    var value Expression
    if p.current.Type == token.ASSIGN {
        p.nextToken()
        value = p.parseExpression()
    }

    // Declara a variável na tabela de símbolos
    if err := p.symbolTable.Declare(nameToken.Lexeme, SymbolInfo{
        Name:      nameToken.Lexeme,
        Type:      typeToken.Lexeme,
        Category:  Variable,
        DefinedAt: nameToken.Line,
    }); err != nil {
        p.addError(err.Error(), nameToken.Line, nameToken.Column)
    }

    return &VariableDeclaration{
        Type:  typeToken.Lexeme,
        Name:  nameToken.Lexeme,
        Value: value,
        Token: nameToken,
    }
}
func isEndOfDeclaration(tok token.Token) bool {
	return tok.Type == token.SEMICOLON ||
		tok.Type == token.EOF ||
		tok.Type == token.TYPE
}
func (p *Parser) parseFunctionDeclaration() *FunctionDeclaration {
    if p.current.Type != token.FUNC {
        p.addError("Expected 'func' keyword", p.current.Line, p.current.Column)
        return nil
    }

    fnToken := p.current
    p.nextToken() // Pula 'func'

    if p.current.Type != token.IDENTIFIER {
        p.addError("Expected function name", p.current.Line, p.current.Column)
        return nil
    }

    name := p.current.Lexeme
    p.nextToken() // Pula nome da função

    // Parênteses de abertura
    if p.current.Type != token.LPAREN {
        p.addError("Expected '(' after function name", p.current.Line, p.current.Column)
        return nil
    }
    p.nextToken() // Pula '('

    var params []*VariableDeclaration
    for p.current.Type != token.RPAREN && !p.AtEnd() {
        // Tipo do parâmetro
        if p.current.Type != token.TYPE {
            p.addError("Expected parameter type", p.current.Line, p.current.Column)
            return nil
        }

        paramType := p.current.Lexeme
        p.nextToken()

        // Nome do parâmetro
        if p.current.Type != token.IDENTIFIER {
            p.addError("Expected parameter name", p.current.Line, p.current.Column)
            return nil
        }

        paramName := p.current.Lexeme
        params = append(params, &VariableDeclaration{
            Type:  paramType,
            Name:  paramName,
            Token: p.current,
        })
        p.nextToken()

        // Verifica separador ou fechamento
        if p.current.Type == token.COMMA {
            p.nextToken()
        } else if p.current.Type != token.RPAREN {
            p.addError(fmt.Sprintf("Expected ',' or ')', got '%s'", p.current.Lexeme), 
                     p.current.Line, p.current.Column)
            return nil
        }
    }

    if p.current.Type != token.RPAREN {
        p.addError("Expected ')' after parameters", p.current.Line, p.current.Column)
        return nil
    }
    p.nextToken() // Pula ')'

    // Tipo de retorno
    if p.current.Type != token.TYPE {
        p.addError("Expected return type", p.current.Line, p.current.Column)
        return nil
    }

    returnType := p.current.Lexeme
    p.nextToken()

    // Corpo da função
    if p.current.Type != token.LBRACE {
        p.addError("Expected '{' to start function body", p.current.Line, p.current.Column)
        return nil
    }

    body := p.parseBlock()
    if body == nil {
        return nil
    }

    return &FunctionDeclaration{
        Name:       name,
        Parameters: params,
        ReturnType: returnType,
        Body:       body.Statements,
        Token:      fnToken,
    }
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
// parser/parser.go
func (p *Parser) Parse() []Statement {
    // Processa todas as funções primeiro
    var funcs []Statement
    for !p.AtEnd() && p.current.Type == token.FUNC {
        funcs = append(funcs, p.parseFunctionDeclaration())
    }
    
    // Processa outras declarações
    for !p.AtEnd() {
        stmt := p.ParseStatement()
        if stmt != nil {
            funcs = append(funcs, stmt)
        }
    }
    return funcs
}
// Função auxiliar para registrar erros
// Em parser/parser.go
func (p *Parser) addError(msg string, line int, column int) {
	p.Errors = append(p.Errors, ParseError{
		Message: msg,
		Line:    line,
		Column:  column,
		Token:   p.current.Lexeme,
	})
}

func (p *Parser) skipUntil(stopTokens ...token.TokenType) {
    for !p.AtEnd() {
        for _, stop := range stopTokens {
            if p.current.Type == stop {
                return
            }
        }
        p.nextToken()
    }
}