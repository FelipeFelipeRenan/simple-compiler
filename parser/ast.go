package parser

import (
	"fmt"

)

// Node representa um nó generico da AST
type Node interface {
	String() string
}

// Expression representa expressoes na AST
type Expression interface {
	Node
	exprNode()
}

// Statement representa comandos como atribuições
type Statement interface {
	Node
	stmtNode()
}

// Identifier representa uma variavel
type Identifier struct {
	Name  string
	Value ValueType
}

func (i *Identifier) exprNode() {}

func (i *Identifier) String() string {
	return i.Name
}

// Number representa um numero na AST
type Number struct {
	Value ValueType
}

func (n *Number) exprNode() {}
func (n *Number) String() string {
	return fmt.Sprintf("%d", n.Value)
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

// Assignment representa uma operação de atribuição
type Assignment struct {
	Variable *Identifier
	Value    Expression
}

func (a *Assignment) stmtNode() {}
func (a *Assignment) String() string {
	return fmt.Sprintf("%s = %s", a.Variable.String(), a.Value.String())
}
