package semantic

import (
	"fmt"
	"simple-compiler/parser"
)

type Analyzer struct {
	ast         []parser.Statement
	symbolTable *parser.SymbolTable
	errors      []SemanticError
}

type SemanticError struct {
	Message string
	Line    int
	Token   string
}

func New(ast []parser.Statement) *Analyzer {
	return &Analyzer{
		ast:         ast,
		symbolTable: parser.NewSymbolTable(),
		errors:      make([]SemanticError, 0),
	}
}

func (a *Analyzer) Analyze() []SemanticError {
	for _, stmt := range a.ast {
		a.checkStatement(stmt)
	}
	return a.errors
}

func (a *Analyzer) addError(msg string, line int, token string) {
	a.errors = append(a.errors, SemanticError{
		Message: msg,
		Line:    line,
		Token:   token,
	})
}

func (a *Analyzer) checkVariableDecl(decl *parser.VariableDeclaration) {
	// Verificação de tipo
	switch decl.Type {
	case "int", "float", "string", "bool":
		// Tipos válidos
	default:
		a.addError(fmt.Sprintf("Tipo desconhecido: %s", decl.Type),
			decl.GetToken().Line, decl.GetToken().Lexeme)
	}
	if a.symbolTable.ExistsInCurrentScope(decl.Name) {
		a.addError(fmt.Sprintf("Variável '%s' já declarada neste escopo", decl.Name),
			decl.Token.Line, decl.Token.Lexeme)
	}

	// Registra a variável
	a.symbolTable.Declare(decl.Name, parser.SymbolInfo{
		Type:      decl.Type,
		Category:  parser.Variable,
		DefinedAt: decl.Token.Line,
	})

	// Verifica a expressão de inicialização
	if decl.Value != nil {
		exprType := a.checkExpression(decl.Value)
		if exprType != "" && !a.isCompatible(decl.Type, exprType) {
			a.addError(fmt.Sprintf("Tipo incompatível: não é possível atribuir %s a %s",
				exprType, decl.Type), decl.Token.Line, decl.Token.Lexeme)
		}
	}
}

func (a *Analyzer) checkAssignment(assign *parser.AssignmentStatement) {
	sym, exists := a.symbolTable.Resolve(assign.Name)
	if !exists {
		a.addError(fmt.Sprintf("Variável '%s' não declarada", assign.Name),
			assign.Token.Line, assign.Token.Lexeme)
		return
	}

	exprType := a.checkExpression(assign.Value)
	if exprType != "" && !a.isCompatible(sym.Type, exprType) {
		a.addError(fmt.Sprintf("Tipo incompatível em atribuição: %s = %s",
			sym.Type, exprType), assign.Token.Line, assign.Token.Lexeme)
	}
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

func (a *Analyzer) checkIdentifier(ident *parser.Identifier) string {
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
		a.checkExpression(s.Expression)
	default:
		a.addError(fmt.Sprintf("Tipo de statement não suportado: %T", stmt),
			stmt.GetToken().Line, stmt.GetToken().Lexeme)
	}
}

func (a *Analyzer) checkIfStatement(ifStmt *parser.IfStatement) {
	condType := a.checkExpression(ifStmt.Condition)
	if condType != "bool" {
		a.addError("Condição do if deve ser booleana",
			ifStmt.Condition.GetToken().Line, ifStmt.Condition.GetToken().Lexeme)
	}

	a.symbolTable.PushScope()
	if ifStmt.Body != nil {
		a.checkBlockStatement(ifStmt.Body)
	}
	a.symbolTable.PopScope()

	if ifStmt.ElseBody != nil {
		a.symbolTable.PushScope()
		a.checkBlockStatement(ifStmt.ElseBody)
		a.symbolTable.PopScope()
	}
}

func (a *Analyzer) checkBinaryExpr(expr *parser.BinaryExpression) string {
	leftType := a.checkExpression(expr.Left)
	rightType := a.checkExpression(expr.Right)

	switch expr.Operator {
	case "+", "-", "*", "/":
		if !a.isNumeric(leftType) || !a.isNumeric(rightType) {
			a.addError(fmt.Sprintf("Operação numérica inválida entre %s e %s",
				leftType, rightType), expr.Token.Line, expr.Token.Lexeme)
			return ""
		}
		return a.resultType(leftType, rightType)

	case ">", "<", ">=", "<=", "==", "!=":
		if !a.isCompatible(leftType, rightType) {
			a.addError(fmt.Sprintf("Comparação inválida entre %s e %s",
				leftType, rightType), expr.Token.Line, expr.Token.Lexeme)
		}
		return "bool"

	case "&&", "||":
		if leftType != "bool" || rightType != "bool" {
			a.addError("Operadores lógicos exigem operandos booleanos",
				expr.Token.Line, expr.Token.Lexeme)
		}
		return "bool"

	default:
		a.addError(fmt.Sprintf("Operador desconhecido: %s", expr.Operator),
			expr.Token.Line, expr.Token.Lexeme)
		return ""
	}
}

func (a *Analyzer) checkBlockStatement(block *parser.BlockStatement) {
	if block == nil {
		return
	}
	a.symbolTable.PushScope()
	for _, stmt := range block.Statements {
		a.checkStatement(stmt)
	}
	a.symbolTable.PopScope()
}

func (a *Analyzer) checkWhileStatement(whileStmt *parser.WhileStatement) {
	condType := a.checkExpression(whileStmt.Condition)
	if condType != "bool" {
		a.addError("Condição do while deve ser booleana",
			whileStmt.Condition.GetToken().Line, whileStmt.Condition.GetToken().Lexeme)
	}

	a.symbolTable.PushScope()
	if whileStmt.Body != nil {
		a.checkBlockStatement(whileStmt.Body)
	}
	a.symbolTable.PopScope()
}

func (a *Analyzer) checkForStatement(forStmt *parser.ForStatement) {
	if forStmt.Init != nil {
		a.checkStatement(forStmt.Init)
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

// Para Identifier