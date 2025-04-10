package intermediatecodegeneration

import (
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
    //case *parser.IfStatement:
      //  cg.generateIfStatement(s)
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