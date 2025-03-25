package parser

import "fmt"

// Node representa um nó genérico da AST
type Node interface {
	String() string
}

// Expression representa expressões na AST
type Expression interface {
	Node
	exprNode()
}

// Statement representa comandos como atribuições
type Statement interface {
	Node
	stmtNode()
}

// Identifier representa uma variável
type Identifier struct {
	Name  string
	Value interface{}
}

func (i *Identifier) exprNode() {}
func (i *Identifier) stmtNode() {}


func (i *Identifier) String() string {
	return fmt.Sprintf("%s", i.Name)
}

// Number representa um número na AST
type Number struct {
	Value float64
}

func (n *Number) exprNode() {}

func (n *Number) String() string {
	return fmt.Sprintf("%v", n.Value)
}

// IfStatement representa uma estrutura condicional
type IfStatement struct {
	Condition Expression
	Body      []Statement
}

func (i *IfStatement) stmtNode() {}

func (i *IfStatement) String() string {
	bodyStr := ""
	for _, stmt := range i.Body {
		bodyStr += "\n    " + stmt.String()
	}
	return fmt.Sprintf("if (%s) {%s\n}", i.Condition.String(), bodyStr)
}

// BinaryExpression representa operações entre dois operandos
type BinaryExpression struct {
	Left     Expression
	Operator string
	Right    Expression
}

func (b *BinaryExpression) exprNode() {}

func (b *BinaryExpression) String() string {
	return fmt.Sprintf("(%s %s %s)", b.Left.String(), b.Operator, b.Right.String())
}

// AssignmentStatement representa uma atribuição de variável na AST
// AssignmentStatement representa uma atribuição de variável
type AssignmentStatement struct {
	Name  string
	Value Expression
}

func (a *AssignmentStatement) stmtNode() {}

func (a *AssignmentStatement) String() string {
	return fmt.Sprintf("(%s = %s)", a.Name, a.Value.String())
}