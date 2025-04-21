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
		if p.peekToken().Type == token.LPAREN {
			return p.parseCallExpression()
		}
		return &Identifier{Name: p.current.Lexeme, Token: p.current}

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

func (p *Parser) parseIfStatement() *IfStatement {
	p.nextToken() // Pula 'if'

	// Parse da condição (com ou sem parênteses)
	var condition Expression
	if p.current.Type == token.LPAREN {
		p.nextToken()
		condition = p.parseExpression()
		if p.current.Type != token.RPAREN {
			p.addError("Esperado ')' após condição", p.current.Line, p.current.Column)
		}
		p.nextToken()
	} else {
		condition = p.parseExpression()
	}

	// Parse do bloco 'then'
	thenBlock := p.parseBlock()

	// Parse do bloco 'else' (opcional)
	var elseBlock *BlockStatement
	if p.current.Type == token.ELSE {
		p.nextToken()
		elseBlock = p.parseBlock()
	}

	return &IfStatement{
		Condition: condition,
		Body:      thenBlock,
		ElseBody:  elseBlock,
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
    block := &BlockStatement{}
    
    if p.current.Type != token.LBRACE {
        return block
    }
    p.nextToken() // Pula '{'

    loopGuard := 0
    maxStatements := 100
    
    for !p.AtEnd() && p.current.Type != token.RBRACE && loopGuard < maxStatements {
        stmt := p.ParseStatement()
        if stmt != nil {
            block.Statements = append(block.Statements, stmt)
        }
        loopGuard++
    }

    if p.current.Type == token.RBRACE {
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
func (p *Parser) addError(msg string, line int, column int) {
	p.Errors = append(p.Errors, ParseError{
		Message: fmt.Sprintf("[Erro] Linha %d:%d: %s (Token: '%s')",
			line, column, msg, p.current.Lexeme),
		Line:   line,
		Column: column,
	})
}

func (p *Parser) skipUntil(stopTypes ...token.TokenType) {
	for !p.AtEnd() && !contains(stopTypes, p.current.Type) {
		p.nextToken()
	}
}

func contains(types []token.TokenType, target token.TokenType) bool {
	for _, t := range types {
		if t == target {
			return true
		}
	}
	return false
}

// Nova função para parsear declaração de função
func (p *Parser) parseFunctionDeclaration() Statement {
    p.nextToken() // Pula 'func'

    // Verificação obrigatória do nome
    if p.current.Type != token.IDENTIFIER {
        p.addError("Nome da função esperado", p.current.Line, p.current.Column)
        return nil
    }
    name := p.current.Lexeme
    p.nextToken()

    // Parâmetros com verificação de progresso
    var params []*VariableDeclaration
    if p.current.Type == token.LPAREN {
        p.nextToken() // Pula '('
        
        paramCounter := 0
        for !p.AtEnd() && p.current.Type != token.RPAREN {
            // Controle anti-loop
            if paramCounter > 10 {
                p.addError("Número máximo de parâmetros excedido", p.current.Line, p.current.Column)
                break
            }

            // Tipo do parâmetro
            if p.current.Type != token.TYPE {
                p.addError("Tipo do parâmetro esperado", p.current.Line, p.current.Column)
                p.nextToken()
                continue
            }
            paramType := p.current.Lexeme
            p.nextToken()

            // Nome do parâmetro
            if p.current.Type != token.IDENTIFIER {
                p.addError("Nome do parâmetro esperado", p.current.Line, p.current.Column)
                p.nextToken()
                continue
            }
            paramName := p.current.Lexeme
            p.nextToken()

            params = append(params, &VariableDeclaration{
                Type: paramType,
                Name: paramName,
            })
            paramCounter++

            // Verifica separador
            if p.current.Type == token.COMMA {
                p.nextToken()
            } else if p.current.Type != token.RPAREN {
                p.addError("Esperado ',' ou ')'", p.current.Line, p.current.Column)
            }
        }
        
        if p.current.Type == token.RPAREN {
            p.nextToken() // Pula ')'
        }
    }

    // Tipo de retorno obrigatório
    returnType := "void"
    if p.current.Type == token.COLON {
        p.nextToken()
        if p.current.Type == token.TYPE {
            returnType = p.current.Lexeme
            p.nextToken()
        } else {
            p.addError("Tipo de retorno esperado", p.current.Line, p.current.Column)
        }
    } else {
        p.addError("Esperado ':' para tipo de retorno", p.current.Line, p.current.Column)
    }

    // Corpo da função com verificação estrita
    if p.current.Type != token.LBRACE {
        p.addError("Esperado '{'", p.current.Line, p.current.Column)
    }
    body := p.parseBlock()
    
    return &FunctionDeclaration{
        Name:       name,
        Parameters: params,
        ReturnType: returnType,
        Body:       body,
    }
}
func (p *Parser) parseCallExpression() Expression {
    ident := &Identifier{Name: p.current.Lexeme}
    p.nextToken() // Pula o nome da função
    
    if p.current.Type != token.LPAREN {
        p.addError("'(' esperado após nome da função", p.current.Line, p.current.Column)
        return nil
    }
    p.nextToken() // Pula '('

    args := []Expression{}
    for p.current.Type != token.RPAREN && !p.AtEnd() {
        arg := p.parseExpression()
        if arg != nil {
            args = append(args, arg)
        }
        
        if p.current.Type == token.COMMA {
            p.nextToken()
        }
    }
    
    if p.current.Type != token.RPAREN {
        p.addError("')' esperado após argumentos", p.current.Line, p.current.Column)
    } else {
        p.nextToken()
    }
    
    return &CallExpression{
        FunctionName: ident.Name,
        Arguments:    args,
    }
}