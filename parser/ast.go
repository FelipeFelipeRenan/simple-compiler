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

func (f *ForStatement) stmtNode() {}
func (f *ForStatement) String() string {
	initStr := ""
	if f.Init != nil {
		initStr = f.Init.String()
	}

	condStr := ""
	if f.Condition != nil {
		condStr = f.Condition.String()
	}

	updateStr := ""
	if f.Update != nil {
		updateStr = f.Update.String()
	}

	bodyStr := ""
	if f.Body != nil {
		for _, stmt := range f.Body.Statements {
			bodyStr += "\n    " + stmt.String()
		}
	}

	return fmt.Sprintf("for (%s; %s; %s) {%s\n}", initStr, condStr, updateStr, bodyStr)
}

// BinaryExpression representa operações entre dois operandos
type BinaryExpression struct {
	Left     Expression
	Operator string
	Right    Expression
	Token    token.Token
}

func (b *BinaryExpression) exprNode() {}
func (b *BinaryExpression) String() string {
	return fmt.Sprintf("(%s %s %s)", b.Left.String(), b.Operator, b.Right.String())
}

// AssignmentStatement representa uma atribuição de variável
type AssignmentStatement struct {
	Name  string
	Value Expression
	Token token.Token
}

func (a *AssignmentStatement) GetToken() token.Token {
	return a.Token
}

func (a *AssignmentStatement) stmtNode() {}
func (a *AssignmentStatement) String() string {
	return fmt.Sprintf("%s = %s", a.Name, a.Value.String())
}

// VariableDeclaration representa declaração de variável
type VariableDeclaration struct {
	Type  string
	Name  string
	Value Expression
	Token token.Token
}

func (v *VariableDeclaration) GetToken() token.Token {
	return v.Token
}

func (vd *VariableDeclaration) stmtNode() {}
func (v *VariableDeclaration) String() string {
	if v.Value != nil {
		return fmt.Sprintf("var %s %s = %s", v.Name, v.Type, v.Value.String())
	}
	return fmt.Sprintf("var %s %s", v.Name, v.Type)
}

// ReturnStatement representa um retorno de função
type ReturnStatement struct {
	Value Expression
}

func (rs *ReturnStatement) stmtNode() {}
func (rs *ReturnStatement) String() string {
	if rs.Value != nil {
		return fmt.Sprintf("return %s", rs.Value.String())
	}
	return "return"
}

// FunctionDeclaration representa uma função
type FunctionDeclaration struct {
	Name       string
	Parameters []*VariableDeclaration
	ReturnType string
	Body       []Statement
}

func (fd *FunctionDeclaration) stmtNode() {}
func (fd *FunctionDeclaration) String() string {
	params := make([]string, len(fd.Parameters))
	for i, p := range fd.Parameters {
		params[i] = p.String()
	}

	body := ""
	for _, stmt := range fd.Body {
		body += "\n    " + stmt.String()
	}

	return fmt.Sprintf("func %s(%s) %s {%s\n}",
		fd.Name, strings.Join(params, ", "), fd.ReturnType, body)
}

type ExpressionStatement struct {
	Expression Expression
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
	return token.Token{}
}

// ReturnStatement
func (r *ReturnStatement) GetToken() token.Token {
	return token.Token{
		Type:   token.RETURN,
		Lexeme: "return",
	}
}

// ExpressionStatement
func (e *ExpressionStatement) GetToken() token.Token {
	return e.Expression.GetToken()
}

func tokenTypeFromOperator(op string) token.TokenType {
	switch op {
	case "+":
		return token.PLUS
	case "-":
		return token.MINUS
	case "*":
		return token.MULT
	case "/":
		return token.DIV
	case ">":
		return token.GT
	case "<":
		return token.LT
	case ">=":
		return token.GTE
	case "<=":
		return token.LTE
	case "==":
		return token.EQ
	case "=":
		return token.ASSIGN
	case "&&":
		return token.AND
	case "||":
		return token.OR
	default:
		return token.ILLEGAL
	}
}

// Implemente para outros tipos conforme necessário
