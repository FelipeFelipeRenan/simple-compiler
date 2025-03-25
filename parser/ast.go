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
	Name string
}

func (i *Identifier) exprNode() {}
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
	ElseBody  []Statement
}

func (i *IfStatement) stmtNode() {}
func (i *IfStatement) String() string {
	bodyStr := ""
	for _, stmt := range i.Body {
		bodyStr += "\n    " + stmt.String()
	}
	elseStr := ""
	if len(i.ElseBody) > 0 {
		elseStr += "\nelse {"
		for _, stmt := range i.ElseBody {
			elseStr += "\n    " + stmt.String()
		}
		elseStr += "\n}"
	}
	return fmt.Sprintf("if (%s) {%s\n}%s", i.Condition.String(), bodyStr, elseStr)
}

// WhileStatement representa um loop while
type WhileStatement struct {
	Condition Expression
	Body      []Statement
}

func (w *WhileStatement) stmtNode() {}
func (w *WhileStatement) String() string {
	bodyStr := ""
	for _, stmt := range w.Body {
		bodyStr += "\n    " + stmt.String()
	}
	return fmt.Sprintf("while (%s) {%s\n}", w.Condition.String(), bodyStr)
}

// ForStatement representa um loop for
type ForStatement struct {
	Init      Statement
	Condition Expression
	Update    Statement
	Body      []Statement
}

func (f *ForStatement) stmtNode() {}
func (f *ForStatement) String() string {
	bodyStr := ""
	for _, stmt := range f.Body {
		bodyStr += "\n    " + stmt.String()
	}
	return fmt.Sprintf("for (%s; %s; %s) {%s\n}", f.Init.String(), f.Condition.String(), f.Update.String(), bodyStr)
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

// AssignmentStatement representa uma atribuição de variável
type AssignmentStatement struct {
	Name  string
	Value Expression
}

func (a *AssignmentStatement) stmtNode() {}
func (a *AssignmentStatement) String() string {
	return fmt.Sprintf("(%s = %s)", a.Name, a.Value.String())
}
