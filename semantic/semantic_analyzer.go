package parser

import (
	"fmt"
	"simple-compiler/token"
	"strings"
)

// Node representa um nó genérico da AST
type Node interface {
	String() string
}

// Expression representa expressões na AST
type Expression interface {
	Node
	exprNode()
	GetToken() token.Token
}

// Statement representa comandos como atribuições
type Statement interface {
	Node
	stmtNode()
	GetToken() token.Token
}

// Identifier representa uma variável
type Identifier struct {
	Name  string
	Token token.Token
}

func (i *Identifier) exprNode() {}
func (i *Identifier) String() string {
	return fmt.Sprintf("%s", i.Name)
}

// Number representa um número na AST
type Number struct {
	Value float64
	Token token.Token
}

func (n *Number) exprNode() {}
func (n *Number) String() string {
	return fmt.Sprintf("%v", n.Value)
}

// IfStatement representa uma estrutura condicional
// IfStatement representa uma estrutura condicional
type IfStatement struct {
	Condition Expression
	Body      *BlockStatement
	ElseBody  *BlockStatement // Opcional
}

func (i *IfStatement) stmtNode() {}
func (i *IfStatement) String() string {
	var sb strings.Builder

	// Condição
	sb.WriteString("if ")
	sb.WriteString(i.Condition.String())
	sb.WriteString(" {\n")

	// Corpo
	if i.Body != nil {
		for _, stmt := range i.Body.Statements {
			sb.WriteString("    ")
			sb.WriteString(stmt.String())
			sb.WriteString("\n")
		}
	}

	sb.WriteString("}")
	return sb.String()
}

// WhileStatement representa um loop while
type WhileStatement struct {
	Condition Expression
	Body      *BlockStatement
}

func (w *WhileStatement) stmtNode() {}
func (w *WhileStatement) String() string {
	if w == nil {
		return "<nil WhileStatement>"
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "while (%s) {", w.Condition.String())

	if w.Body != nil {
		for _, stmt := range w.Body.Statements {
			sb.WriteString("\n" + indent(stmt.String()))
		}
	}

	sb.WriteString("\n}")
	return sb.String()
}

// ForStatement representa um loop for
type ForStatement struct {
	Init      Statement  // Declaração ou atribuição
	Condition Expression // Expressão booleana
	Update    Statement  // Atribuição
	Body      *BlockStatement
}

func (a *Analyzer) isCompatible(targetType, exprType string) bool {
	compatibility := map[string]map[string]bool{
		"int":    {"int": true, "float": true},
		"float":  {"float": true, "int": true},
		"string": {"string": true},
		"bool":   {"bool": true},
	}

	if rules, ok := compatibility[targetType]; ok {
		return rules[exprType]
	}
	return false
}

func (a *Analyzer) isNumeric(typeName string) bool {
	return typeName == "int" || typeName == "float"
}

func (a *Analyzer) resultType(leftType, rightType string) string {
	if leftType == "float" || rightType == "float" {
		return "float"
	}
	return "int"
}


func (a *Analyzer) checkExpression(expr parser.Expression) string {
	switch e := expr.(type) {

	case *parser.Identifier:
		return a.checkIdentifier(e)
	case *parser.BinaryExpression:
		return a.checkBinaryExpr(e)
	case *parser.Number:
		if e.Value == float64(int(e.Value)) {
			return "int"
		}
		return "float"
	case *parser.BooleanLiteral:
		return "bool"
	case *parser.StringLiteral:
		return "string"
	default:
		a.addError(fmt.Sprintf("Tipo de expressão não suportado: %T", expr),
			expr.GetToken().Line, expr.GetToken().Lexeme)
		return ""
	}
}

// Modifique a verificação de identificador
func (a *Analyzer) checkIdentifier(ident *parser.Identifier) string {
	// Verifica se é uma função builtin
	if a.isBuiltinFunction(ident.Name) {
		return "void" // print não retorna valor
	}

	sym, exists := a.symbolTable.Resolve(ident.Name)
	if !exists {
		a.addError(fmt.Sprintf("Identificador não declarado: %s", ident.Name),
			ident.Token.Line, ident.Token.Lexeme)
		return ""
	}
	return sym.Type
}

func (a *Analyzer) checkStatement(stmt parser.Statement) {
	switch s := stmt.(type) {
	case *parser.VariableDeclaration:
		a.checkVariableDecl(s)
	case *parser.AssignmentStatement:
		a.checkAssignment(s)
	case *parser.IfStatement:
		a.checkIfStatement(s)
	case *parser.WhileStatement:
		a.checkWhileStatement(s)
	case *parser.ForStatement:
		a.checkForStatement(s)
	case *parser.BlockStatement:
		a.checkBlockStatement(s)
    case *parser.ExpressionStatement:
        if call, ok := s.Expression.(*parser.CallExpression); ok && call.FunctionName == "print" {
            a.checkPrintCall(call)
        } else {
            a.checkExpression(s.Expression)
        }

	case *parser.FunctionDeclaration:
		a.symbolTable.PushScope()
		// Verifica parâmetros
		for _, param := range s.Parameters {
			a.checkVariableDecl(param)
		}
		// Verifica corpo
		for _, stmt := range s.Body {
			a.checkStatement(stmt)
		}
		a.symbolTable.PopScope()
		a.checkFunctionDecl(s)
	default:
		a.addError(fmt.Sprintf("Tipo de statement não suportado: %T", stmt),
			stmt.GetToken().Line, stmt.GetToken().Lexeme)
	}
}

func (es *ExpressionStatement) stmtNode() {}
func (es *ExpressionStatement) String() string {
	return es.Expression.String()
}

// Função auxiliar para indentação
func indent(s string) string {
	return "    " + strings.ReplaceAll(s, "\n", "\n    ")
}

type BlockStatement struct {
	Statements []Statement
}

func (b *BlockStatement) stmtNode() {}
func (b *BlockStatement) String() string {
	var sb strings.Builder
	for _, stmt := range b.Statements {
		sb.WriteString(stmt.String())
	}
	return sb.String()
}

type BooleanLiteral struct {
	Value bool
	Token token.Token
}

func (b *BooleanLiteral) exprNode()             {}
func (b *BooleanLiteral) String() string        { return fmt.Sprintf("%v", b.Value) }
func (b *BooleanLiteral) GetToken() token.Token { return b.Token }

type StringLiteral struct {
	Value string
	Token token.Token
}

func (s *StringLiteral) exprNode()             {}
func (s *StringLiteral) String() string        { return fmt.Sprintf("\"%s\"", s.Value) }
func (s *StringLiteral) GetToken() token.Token { return s.Token }

// BinaryExpression
func (b *BinaryExpression) GetToken() token.Token {
	return token.Token{
		Type:   tokenTypeFromOperator(b.Operator),
		Lexeme: b.Operator,
	}
}

type UnaryExpression struct {
    Operator string
    Right    Expression
    Token    token.Token
}

func (u *UnaryExpression) exprNode() {}
func (u *UnaryExpression) String() string {
    return fmt.Sprintf("(%s%s)", u.Operator, u.Right.String())
}
func (u *UnaryExpression) GetToken() token.Token {
    return u.Token
}

// Number
func (n *Number) GetToken() token.Token {
	lexeme := fmt.Sprintf("%v", n.Value)
	if n.Value == float64(int(n.Value)) {
		lexeme = fmt.Sprintf("%d", int(n.Value))
	}
	return token.Token{
		Type:   token.NUMBER,
		Lexeme: lexeme,
	}
}

// Identifier
func (i *Identifier) GetToken() token.Token {
	return token.Token{
		Type:   token.IDENTIFIER,
		Lexeme: i.Name,
	}
}

// IfStatement
func (i *IfStatement) GetToken() token.Token {
	return token.Token{
		Type:   token.IF,
		Lexeme: "if",
	}
}

// WhileStatement
func (w *WhileStatement) GetToken() token.Token {
	return token.Token{
		Type:   token.WHILE,
		Lexeme: "while",
	}
}

// ForStatement
func (f *ForStatement) GetToken() token.Token {
	return token.Token{
		Type:   token.FOR,
		Lexeme: "for",
	}
}

// BlockStatement
func (b *BlockStatement) GetToken() token.Token {
	if len(b.Statements) > 0 {
		return b.Statements[0].GetToken()
	}

	if forStmt.Condition != nil {
		condType := a.checkExpression(forStmt.Condition)
		if condType != "bool" {
			a.addError("Condição do for deve ser booleana",
				forStmt.Condition.GetToken().Line, forStmt.Condition.GetToken().Lexeme)
		}
	}

	if forStmt.Update != nil {
		a.checkStatement(forStmt.Update)
	}

	a.symbolTable.PushScope()
	if forStmt.Body != nil {
		a.checkBlockStatement(forStmt.Body)
	}
	a.symbolTable.PopScope()
}

// semantic/semantic_analyzer.go

// semantic/semantic_analyzer.go
func (a *Analyzer) checkFunctionDecl(fd *parser.FunctionDeclaration) {
	// Registra função no escopo GLOBAL
	a.symbolTable.Declare(fd.Name, parser.SymbolInfo{
		Name:      fd.Name,
		Type:      fd.ReturnType,
		Category:  parser.Function,
		DefinedAt: fd.Token.Line,
	})

	// Cria escopo LOCAL para parâmetros
	a.symbolTable.PushScope()

	// Registra parâmetros
	for _, param := range fd.Parameters {
		a.symbolTable.Declare(param.Name, parser.SymbolInfo{
			Name:      param.Name,
			Type:      param.Type,
			Category:  parser.Variable,
			DefinedAt: param.Token.Line,
		})
	}

	// Verifica corpo
	for _, stmt := range fd.Body {
		a.checkStatement(stmt)
	}

	a.symbolTable.PopScope()
}

func (a *Analyzer) isBuiltinFunction(name string) bool {
	return name == "print"
}

func (a *Analyzer) checkCallExpression(call *parser.CallExpression) string {
	if call.FunctionName == "print" {
		if len(call.Arguments) != 1 {
			a.addError("print requer exatamente 1 argumento", call.Token.Line, call.Token.Lexeme)
		} else {
			argType := a.checkExpression(call.Arguments[0])
			if argType != "int" && argType != "float" {
				a.addError(fmt.Sprintf("print só suporta int ou float, recebeu %s", argType),
					call.Token.Line, call.Token.Lexeme)
			}
		}
		return "void"
	}
	// ... resto da implementação original
	return ""
}


func (a *Analyzer) checkPrintCall(call *parser.CallExpression) {
    if len(call.Arguments) != 1 {
        a.addError("print requer exatamente 1 argumento", call.Token.Line, call.Token.Lexeme)
        return
    }
    
    argType := a.checkExpression(call.Arguments[0])
    if argType != "int" && argType != "float" {
        a.addError(fmt.Sprintf("print só suporta int ou float, recebeu %s", argType),
            call.Token.Line, call.Token.Lexeme)
    }
}