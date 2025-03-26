package semantic

import (
	"fmt"
	"simple-compiler/parser"
)

type Analyzer struct {
	ast         []parser.Statement
	symbolTable *parser.SymbolTable
	errors      []string
}

func New(ast []parser.Statement) *Analyzer {
	return &Analyzer{
		ast:         ast,
		symbolTable: parser.NewSymbolTable(),
		errors:      make([]string, 0),
	}

}

func (a *Analyzer) Analyze() []string{
	for _, stmt := range a.ast {
		a.checkStatement(stmt)
	}
	return a.errors
}
func (a *Analyzer) checkVariableDecl(decl *parser.VariableDeclaration) {
    if a.symbolTable.ExistsInCurrentScope(decl.Name) {
        a.errors = append(a.errors, fmt.Sprintf(
            "Variável '%s' já declarada neste escopo", decl.Name))
    }

    // Registra a variável
    a.symbolTable.Declare(decl.Name, parser.SymbolInfo{
        Type:     decl.Type,
        Category: parser.Variable,
    })

    // Verifica a expressão de inicialização
    if decl.Value != nil {
        exprType := a.checkExpression(decl.Value)
        if exprType != "" && !a.isCompatible(decl.Type, exprType) {
            a.errors = append(a.errors, fmt.Sprintf(
                "Tipo incompatível: não é possível atribuir %s a %s",
                exprType, decl.Type))
        }
    }
}

func (a *Analyzer) checkAssignment(assign *parser.AssignmentStatement) {
    sym, exists := a.symbolTable.Resolve(assign.Name)
    if !exists {
        a.errors = append(a.errors, fmt.Sprintf(
            "Variável '%s' não declarada", assign.Name))
        return
    }

    exprType := a.checkExpression(assign.Value)
    if exprType != "" && !a.isCompatible(sym.Type, exprType) {
        a.errors = append(a.errors, fmt.Sprintf(
            "Tipo incompatível em atribuição: %s = %s",
            sym.Type, exprType))
    }
}
// Adicione essas funções dentro do tipo Analyzer
func (a *Analyzer) isCompatible(targetType, exprType string) bool {
    // Tabela de compatibilidade de tipos
    compatibility := map[string]map[string]bool{
        "int": {
            "int":   true,
            "float": true,
        },
        "float": {
            "float": true,
            "int":   true,
        },
        "string": {
            "string": true,
        },
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
    default:
        a.errors = append(a.errors, fmt.Sprintf("Tipo de expressão não suportado: %T", expr))
        return ""
    }
}

func (a *Analyzer) checkIdentifier(ident *parser.Identifier) string {
    sym, exists := a.symbolTable.Resolve(ident.Name)
    if !exists {
        a.errors = append(a.errors, fmt.Sprintf("Identificador não declarado: %s", ident.Name))
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
    case *parser.BlockStatement:
        a.checkBlockStatement(s)
    default:
        a.errors = append(a.errors, 
            fmt.Sprintf("Tipo de statement não suportado: %T", stmt))
    }
}
func (a *Analyzer) checkIfStatement(ifStmt *parser.IfStatement) {
    // Verifica a condição
    condType := a.checkExpression(ifStmt.Condition)
    if condType != "bool" {
        a.errors = append(a.errors, "Condição do if deve ser booleana")
    }

    // Verifica o bloco principal
    a.symbolTable.PushScope()
    if ifStmt.Body != nil {
        for _, stmt := range ifStmt.Body.Statements {  // Acesse Statements dentro do BlockStatement
            a.checkStatement(stmt)
        }
    }
    a.symbolTable.PopScope()

    // Verifica o bloco else
    if ifStmt.ElseBody != nil {
        a.symbolTable.PushScope()
        for _, stmt := range ifStmt.ElseBody.Statements {  // Acesse Statements dentro do BlockStatement
            a.checkStatement(stmt)
        }
        a.symbolTable.PopScope()
    }
}

func (a *Analyzer) checkBinaryExpr(expr *parser.BinaryExpression) string {
    leftType := a.checkExpression(expr.Left)
    rightType := a.checkExpression(expr.Right)

    // Verificação de operações
    switch expr.Operator {
    case "+", "-", "*", "/":
        if !a.isNumeric(leftType) || !a.isNumeric(rightType) {
            a.errors = append(a.errors, fmt.Sprintf(
                "Operação numérica inválida entre %s e %s", 
                leftType, rightType))
            return ""
        }
        return a.resultType(leftType, rightType)
    
    case ">", "<", ">=", "<=", "==", "!=":
        if !a.isCompatible(leftType, rightType) {
            a.errors = append(a.errors, fmt.Sprintf(
                "Comparação inválida entre %s e %s", 
                leftType, rightType))
        }
        return "bool"
    
    default:
        a.errors = append(a.errors, 
            fmt.Sprintf("Operador desconhecido: %s", expr.Operator))
        return ""
    }
}

func (a *Analyzer) checkBlockStatement(block *parser.BlockStatement) {
    if block == nil {
        return
    }
    for _, stmt := range block.Statements {
        a.checkStatement(stmt)
    }
}

func (a *Analyzer) checkWhileStatement(whileStmt *parser.WhileStatement) {
    // Verifica a condição
    condType := a.checkExpression(whileStmt.Condition)
    if condType != "bool" {
        a.errors = append(a.errors, "Condição do while deve ser booleana")
    }

    // Verifica o corpo
    a.symbolTable.PushScope()
    if whileStmt.Body != nil {
        for _, stmt := range whileStmt.Body.Statements {  // Acesse Statements dentro do BlockStatement
            a.checkStatement(stmt)
        }
    }
    a.symbolTable.PopScope()
}