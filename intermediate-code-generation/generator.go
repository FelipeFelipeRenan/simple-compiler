package intermediatecodegeneration

import (
	"fmt"
	"simple-compiler/parser"
	"strconv"
	"strings"
)

type CodeGenerator struct {
	ir           *IntermediateRep
	labelCounter int
	tempCounter  int
}

func NewCodeGenerator() *CodeGenerator {
	return &CodeGenerator{
		ir:           NewIR(),
		labelCounter: 0,
		tempCounter:  0,
	}
}

func (cg *CodeGenerator) NewLabel() string {
	label := fmt.Sprintf("L%d", cg.labelCounter)
	cg.labelCounter++
	return label
}

func (cg *CodeGenerator) NewTemp() string {
	temp := fmt.Sprintf("t%d", cg.tempCounter)
	cg.tempCounter++
	return temp
}

func (cg *CodeGenerator) GenerateFromAST(statements []parser.Statement) *IntermediateRep {
	for _, stmt := range statements {
		cg.generateStatement(stmt)
	}
	return cg.ir
}

func (cg *CodeGenerator) generateStatement(stmt parser.Statement) {
    switch s := stmt.(type) {
    case *parser.VariableDeclaration:
        cg.generateVariableDecl(s)
    case *parser.AssignmentStatement:
        cg.generateAssignment(s)
    case *parser.IfStatement:
        cg.generateIfStatement(s)
    case *parser.WhileStatement:
        cg.generateWhileStatement(s)
    case *parser.ForStatement:
        cg.generateForStatement(s)
    case *parser.ReturnStatement:
        cg.generateReturnStatement(s)
    case *parser.BlockStatement:
        cg.generateBlock(s)
    case *parser.ExpressionStatement:
        cg.generateExpression(s.Expression)
    default:
        fmt.Printf("Tipo de statement não suportado: %T\n", s)
    }
}

func (cg *CodeGenerator) generateVariableDecl(decl *parser.VariableDeclaration) {
	if decl.Value != nil {
		temp := cg.generateExpression(decl.Value)
		cg.ir.Instructions = append(cg.ir.Instructions, Instruction{
			Op:   ASSIGN,
			Dest: decl.Name,
			Arg1: temp,
		})
	}
}

func (cg *CodeGenerator) generateAssignment(assign *parser.AssignmentStatement) {
	temp := cg.generateExpression(assign.Value)
	cg.ir.Instructions = append(cg.ir.Instructions, Instruction{
		Op:   ASSIGN,
		Dest: assign.Name,
		Arg1: temp,
	})
}

func (cg *CodeGenerator) generateExpression(expr parser.Expression) string {
    switch e := expr.(type) {
    case *parser.Identifier:
        return e.Name
    case *parser.Number:
        return strconv.FormatFloat(e.Value, 'f', -1, 64)
    case *parser.BinaryExpression:
        return cg.generateBinaryExpr(e)
    case *parser.BooleanLiteral:
        if e.Value { return "1" }
        return "0"
    case *parser.UnaryExpression:
        return cg.generateUnaryExpr(e)
    case *parser.CallExpression:
        return cg.generateCallExpr(e)
    default:
        fmt.Printf("Tipo de expressão não suportado: %T\n", e)
        return ""
    }
}


func (cg *CodeGenerator) generateBinaryExpr(expr *parser.BinaryExpression) string {
	left := cg.generateExpression(expr.Left)
	right := cg.generateExpression(expr.Right)
	temp := cg.NewTemp()

	cg.ir.Instructions = append(cg.ir.Instructions, Instruction{
		Op:   Operation(expr.Operator),
		Dest: temp,
		Arg1: left,
		Arg2: right,
	})

	return temp
}

func (cg *CodeGenerator) generateIfStatement(ifStmt *parser.IfStatement) {
	// Gera código para a condição
	condTemp := cg.generateExpression(ifStmt.Condition)

	// Cria labels
	elseLabel := cg.NewLabel()
	endLabel := cg.NewLabel()

	// Gera o jump condicional
	cg.ir.Instructions = append(cg.ir.Instructions, Instruction{
		Op:    "if_false",
		Arg1:  condTemp,
		Label: elseLabel,
	})

	// Gera o bloco THEN
	cg.generateBlock(ifStmt.Body)

	// Gera o jump para o final
	cg.ir.Instructions = append(cg.ir.Instructions, Instruction{
		Op:    "goto",
		Label: endLabel,
	})

	// Gera o label do ELSE
	cg.ir.Instructions = append(cg.ir.Instructions, Instruction{
		Op:    "label",
		Label: elseLabel,
	})

	// Gera o bloco ELSE (se existir)
	if ifStmt.ElseBody != nil {
		cg.generateBlock(ifStmt.ElseBody)
	}

	// Gera o label de fim
	cg.ir.Instructions = append(cg.ir.Instructions, Instruction{
		Op:    "label",
		Label: endLabel,
	})
}

func (cg *CodeGenerator) generateBlock(block *parser.BlockStatement) {
    if block == nil {
        return
    }
    for _, stmt := range block.Statements {
        cg.generateStatement(stmt)
    }
}

func (cg *CodeGenerator) generateWhileStatement(whileStmt *parser.WhileStatement) {
	startLabel := fmt.Sprintf("L%d", cg.labelCounter)
	endLabel := fmt.Sprintf("L%d", cg.labelCounter+1)
	cg.labelCounter += 2

	// Label de início do loop
	cg.ir.Instructions = append(cg.ir.Instructions, Instruction{
		Op:    "label",
		Label: startLabel,
	})

	// Avalia condição
	condTemp := cg.generateExpression(whileStmt.Condition)

	// Sai do loop se condição falsa
	cg.ir.Instructions = append(cg.ir.Instructions, Instruction{
		Op:    "if_false",
		Arg1:  condTemp,
		Label: endLabel,
	})

	// Corpo do loop
	cg.generateBlock(whileStmt.Body)

	// Volta para o início
	cg.ir.Instructions = append(cg.ir.Instructions, Instruction{
		Op:    "goto",
		Label: startLabel,
	})

	// Label de saída
	cg.ir.Instructions = append(cg.ir.Instructions, Instruction{
		Op:    "label",
		Label: endLabel,
	})
}

func (cg *CodeGenerator) generateLogicalOp(expr *parser.BinaryExpression) string {
	arg1 := cg.generateExpression(expr.Left)
	arg2 := cg.generateExpression(expr.Right)
	temp := cg.ir.NewTemp()

	op := ""
	switch expr.Operator {
	case "&&":
		op = "and"
	case "||":
		op = "or"
	}

	cg.ir.Instructions = append(cg.ir.Instructions, Instruction{
		Op:   Operation(op),
		Dest: temp,
		Arg1: arg1,
		Arg2: arg2,
	})

	return temp
}

func (cg *CodeGenerator) generateUnaryExpr(expr *parser.UnaryExpression) string {
    right := cg.generateExpression(expr.Right)
    temp := cg.NewTemp()
    
    var op Operation
    switch expr.Operator {
    case "-":
        op = "neg"
    case "!":
        op = "not"
    default:
        op = Operation(expr.Operator)
    }
    
    cg.ir.Instructions = append(cg.ir.Instructions, Instruction{
        Op:   op,
        Dest: temp,
        Arg1: right,
    })
    
    return temp
}

func (cg *CodeGenerator) generateCallExpr(call *parser.CallExpression) string {
    temp := cg.NewTemp()
    args := make([]string, 0, len(call.Arguments))
    
    for _, arg := range call.Arguments {
        args = append(args, cg.generateExpression(arg))
    }
    
    cg.ir.Instructions = append(cg.ir.Instructions, Instruction{
        Op:    "call",
        Dest:  temp,
        Arg1:  call.FunctionName,
        Arg2:  strings.Join(args, ","), // Argumentos como string separada por vírgulas
    })
    
    return temp
}
func (cg *CodeGenerator) generateReturnStatement(ret *parser.ReturnStatement) {
	if ret.Value != nil {
		val := cg.generateExpression(ret.Value)
		cg.ir.Instructions = append(cg.ir.Instructions, Instruction{
			Op:   "return",
			Arg1: val,
		})
	} else {
		cg.ir.Instructions = append(cg.ir.Instructions, Instruction{
			Op: "return",
		})
	}
}

func (cg *CodeGenerator) generateForStatement(forStmt *parser.ForStatement) {
	startLabel := cg.NewLabel()
	endLabel := cg.NewLabel()

	// Inicialização
	if forStmt.Init != nil {
		cg.generateStatement(forStmt.Init)
	}

	// Label de início
	cg.ir.Instructions = append(cg.ir.Instructions, Instruction{
		Op:    "label",
		Label: startLabel,
	})

	// Condição
	if forStmt.Condition != nil {
		cond := cg.generateExpression(forStmt.Condition)
		cg.ir.Instructions = append(cg.ir.Instructions, Instruction{
			Op:    "if_false",
			Arg1:  cond,
			Label: endLabel,
		})
	}

	// Corpo do loop
	cg.generateBlock(forStmt.Body)

	// Atualização
	if forStmt.Update != nil {
		cg.generateStatement(forStmt.Update)
	}

	// Volta para o início
	cg.ir.Instructions = append(cg.ir.Instructions, Instruction{
		Op:    "goto",
		Label: startLabel,
	})

	// Label de fim
	cg.ir.Instructions = append(cg.ir.Instructions, Instruction{
		Op:    "label",
		Label: endLabel,
	})
}
