package parser

import (
	"fmt"
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
    return fmt.Sprintf("%s = %s", a.Name, a.Value.String())
}


// VariableDeclaration representa declaração de variável
type VariableDeclaration struct {
    Type  string
    Name  string
    Value Expression
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