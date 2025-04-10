package intermediatecodegeneration

import (
	"fmt"
	"simple-compiler/parser"
	"strconv"
)

type CodeGenerator struct {
	ir *IntermediateRep
}

func NewCodeGenerator() *CodeGenerator {
	return &CodeGenerator{
		ir: NewIR(),
	}
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
	case *parser.BlockStatement:
		cg.generateBlock(s)

		// ... outros casos
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
		// ... outros casos
	}
	return ""
}

func (cg *CodeGenerator) generateBinaryExpr(expr *parser.BinaryExpression) string {
	arg1 := cg.generateExpression(expr.Left)
	arg2 := cg.generateExpression(expr.Right)
	temp := cg.ir.NewTemp()

	cg.ir.Instructions = append(cg.ir.Instructions, Instruction{
		Op:   Operation(expr.Operator),
		Dest: temp,
		Arg1: arg1,
		Arg2: arg2,
	})

	return temp
}

func (cg *CodeGenerator) generateIfStatement(ifStmt *parser.IfStatement) {
	// Gera código para a condição
	condTemp := cg.generateExpression(ifStmt.Condition)

	// Cria labels para os jumps
	elseLabel := fmt.Sprintf("L%d", cg.ir.TempCount)
	endLabel := fmt.Sprintf("L%d", cg.ir.TempCount+1)
	cg.ir.TempCount += 2

	// Adiciona instrução condicional
	cg.ir.Instructions = append(cg.ir.Instructions, Instruction{
		Op:    IFLT, // Você pode ajustar o operador conforme necessário
		Arg1:  condTemp,
		Arg2:  "0", // Compara com zero (false)
		Label: elseLabel,
	})

	// Gera código para o bloco then
	cg.generateBlock(ifStmt.Body)

	// Adiciona jump para o final (para pular o else)
	cg.ir.Instructions = append(cg.ir.Instructions, Instruction{
		Op:    GOTO,
		Label: endLabel,
	})

	// Adiciona label para o else
	cg.ir.Instructions = append(cg.ir.Instructions, Instruction{
		Op:    LABEL,
		Label: elseLabel,
	})

	// Gera código para o bloco else (se existir)
	if ifStmt.ElseBody != nil {
		cg.generateBlock(ifStmt.ElseBody)
	}

	// Adiciona label para o final
	cg.ir.Instructions = append(cg.ir.Instructions, Instruction{
		Op:    LABEL,
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
