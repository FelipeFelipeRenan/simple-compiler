package intermediatecodegeneration

import (
	"fmt"
	"simple-compiler/parser"
	"strconv"
)

type CodeGenerator struct {
	ir           *IntermediateRep
	symbolTable  map[string]VariableInfo
	currentBlock *BasicBlock
	tempCounter  int
	labelCounter int
}

type VariableInfo struct {
	Alloca string
	Type   Type
}

func NewCodeGenerator() *CodeGenerator {
	ir := NewIR()
	return &CodeGenerator{
		ir:           ir,
		symbolTable:  make(map[string]VariableInfo),
		currentBlock: ir.CurrentBlock(),
		tempCounter:  0,
		labelCounter: 0,
	}
}

func (cg *CodeGenerator) GenerateFromAST(statements []parser.Statement) *IntermediateRep {
	for _, stmt := range statements {
		cg.generateStatement(stmt)
	}

	if cg.currentBlock.Terminator == nil {
		cg.currentBlock.Terminator = &Instruction{
			Op:   "ret",
			Type: I32,
			Args: []string{"0"},
		}
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
	}
}

func (cg *CodeGenerator) generateVariableDecl(decl *parser.VariableDeclaration) {
	llvmType := cg.llvmTypeFromParserType(decl.Type)
	alloca := cg.newTemp()

	cg.currentBlock.Instructions = append(cg.currentBlock.Instructions, Instruction{
		Op:   "alloca",
		Type: llvmType,
		Dest: alloca,
	})

	cg.symbolTable[decl.Name] = VariableInfo{
		Alloca: alloca,
		Type:   llvmType,
	}

	if decl.Value != nil {
		val := cg.generateExpression(decl.Value)
		cg.currentBlock.Instructions = append(cg.currentBlock.Instructions, Instruction{
			Op:   "store",
			Type: llvmType,
			Args: []string{val, fmt.Sprintf("%s*", llvmType), alloca},
		})
	}
}

func (cg *CodeGenerator) generateAssignment(assign *parser.AssignmentStatement) {
	info, exists := cg.symbolTable[assign.Name]
	if !exists {
		return
	}

	val := cg.generateExpression(assign.Value)
	cg.currentBlock.Instructions = append(cg.currentBlock.Instructions, Instruction{
		Op:   "store",
		Type: info.Type,
		Args: []string{val, fmt.Sprintf("%s*", info.Type), info.Alloca},
	})
}

func (cg *CodeGenerator) generateExpression(expr parser.Expression) string {
	switch e := expr.(type) {
	case *parser.Identifier:
		return cg.generateIdentifier(e)
	case *parser.Number:
		return cg.generateNumber(e)
	case *parser.BinaryExpression:
		return cg.generateBinaryExpr(e)
	case *parser.BooleanLiteral:
		return cg.generateBooleanLiteral(e)
	case *parser.UnaryExpression:
		return cg.generateUnaryExpr(e)
	case *parser.CallExpression:
		return cg.generateCallExpr(e)
	default:
		return "0"
	}
}

func (cg *CodeGenerator) generateIdentifier(ident *parser.Identifier) string {
	info, exists := cg.symbolTable[ident.Name]
	if !exists {
		return "0"
	}

	temp := cg.newTemp()
	cg.currentBlock.Instructions = append(cg.currentBlock.Instructions, Instruction{
		Op:   "load",
		Type: info.Type,
		Args: []string{fmt.Sprintf("%s*", info.Type), info.Alloca},
		Dest: temp,
	})
	return temp
}

func (cg *CodeGenerator) generateNumber(num *parser.Number) string {
	if num.Value == float64(int(num.Value)) {
		return strconv.Itoa(int(num.Value))
	}
	return fmt.Sprintf("%f", num.Value)
}

func (cg *CodeGenerator) generateBooleanLiteral(boolLit *parser.BooleanLiteral) string {
	if boolLit.Value {
		return "1"
	}
	return "0"
}

func (cg *CodeGenerator) generateUnaryExpr(expr *parser.UnaryExpression) string {
	right := cg.generateExpression(expr.Right)
	temp := cg.newTemp()

	switch expr.Operator {
	case "-":
		typ := cg.determineType(expr.Right)
		if typ == FLOAT {
			cg.currentBlock.Instructions = append(cg.currentBlock.Instructions, Instruction{
				Op:   "fneg",
				Type: FLOAT,
				Dest: temp,
				Args: []string{right},
			})
		} else {
			cg.currentBlock.Instructions = append(cg.currentBlock.Instructions, Instruction{
				Op:   "sub",
				Type: I32,
				Dest: temp,
				Args: []string{"0", right},
			})
		}
	case "!":
		cg.currentBlock.Instructions = append(cg.currentBlock.Instructions, Instruction{
			Op:   "xor",
			Type: I1,
			Dest: temp,
			Args: []string{right, "1"},
		})
	default:
		return right
	}

	return temp
}

func (cg *CodeGenerator) generateBinaryExpr(expr *parser.BinaryExpression) string {
	left := cg.generateExpression(expr.Left)
	right := cg.generateExpression(expr.Right)
	temp := cg.newTemp()

	leftType := cg.determineType(expr.Left)
	rightType := cg.determineType(expr.Right)
	resultType := leftType

	if leftType == FLOAT || rightType == FLOAT {
		resultType = FLOAT
		if leftType != FLOAT {
			left = cg.generateTypeConversion(left, leftType, FLOAT)
		}
		if rightType != FLOAT {
			right = cg.generateTypeConversion(right, rightType, FLOAT)
		}
	}

	var op string
	switch expr.Operator {
	case "+":
		if resultType == FLOAT {
			op = "fadd"
		} else {
			op = "add"
		}
	case "-":
		if resultType == FLOAT {
			op = "fsub"
		} else {
			op = "sub"
		}
	case "*":
		if resultType == FLOAT {
			op = "fmul"
		} else {
			op = "mul"
		}
	case "/":
		if resultType == FLOAT {
			op = "fdiv"
		} else {
			op = "sdiv"
		}
	case "<", ">", "<=", ">=", "==", "!=":
		return cg.generateComparison(expr, left, right, leftType, rightType)
	default:
		return "0"
	}

	cg.currentBlock.Instructions = append(cg.currentBlock.Instructions, Instruction{
		Op:   op,
		Type: resultType,
		Dest: temp,
		Args: []string{left, right},
	})

	return temp
}

func (cg *CodeGenerator) generateComparison(expr *parser.BinaryExpression, left, right string, leftType, rightType Type) string {
	temp := cg.newTemp()
	var op string

	if leftType == FLOAT || rightType == FLOAT {
		switch expr.Operator {
		case "<": op = "fcmp olt"
		case ">": op = "fcmp ogt"
		case "<=": op = "fcmp ole"
		case ">=": op = "fcmp oge"
		case "==": op = "fcmp oeq"
		case "!=": op = "fcmp one"
		}
	} else {
		switch expr.Operator {
		case "<": op = "icmp slt"
		case ">": op = "icmp sgt"
		case "<=": op = "icmp sle"
		case ">=": op = "icmp sge"
		case "==": op = "icmp eq"
		case "!=": op = "icmp ne"
		}
	}

	cg.currentBlock.Instructions = append(cg.currentBlock.Instructions, Instruction{
		Op:   op,
		Type: I1,
		Dest: temp,
		Args: []string{left, right},
	})

	return temp
}

func (cg *CodeGenerator) generateTypeConversion(value string, fromType, toType Type) string {
	temp := cg.newTemp()
	var op string

	if fromType == I32 && toType == FLOAT {
		op = "sitofp"
	} else if fromType == FLOAT && toType == I32 {
		op = "fptosi"
	} else {
		return value
	}

	cg.currentBlock.Instructions = append(cg.currentBlock.Instructions, Instruction{
		Op:   op,
		Type: toType,
		Dest: temp,
		Args: []string{value},
	})

	return temp
}

func (cg *CodeGenerator) generateCallExpr(call *parser.CallExpression) string {
	temp := cg.newTemp()
	args := make([]string, len(call.Arguments))

	for i, arg := range call.Arguments {
		args[i] = cg.generateExpression(arg)
	}

	cg.currentBlock.Instructions = append(cg.currentBlock.Instructions, Instruction{
		Op:   "call",
		Dest: temp,
		Args: append([]string{call.FunctionName}, args...),
	})

	return temp
}

func (cg *CodeGenerator) generateIfStatement(ifStmt *parser.IfStatement) {
	cond := cg.generateExpression(ifStmt.Condition)
	thenLabel := cg.newLabel("if.then")
	elseLabel := cg.newLabel("if.else")
	endLabel := cg.newLabel("if.end")

	cg.currentBlock.Terminator = &Instruction{
		Op:   "br",
		Args: []string{cond, thenLabel, elseLabel},
	}

	// Then block
	thenBlock := &BasicBlock{Label: thenLabel}
	cg.ir.CurrentFunction().Blocks = append(cg.ir.CurrentFunction().Blocks, thenBlock)
	cg.currentBlock = thenBlock
	cg.generateBlock(ifStmt.Body)
	thenBlock.Terminator = &Instruction{
		Op:   "br",
		Args: []string{endLabel},
	}

	// Else block
	elseBlock := &BasicBlock{Label: elseLabel}
	cg.ir.CurrentFunction().Blocks = append(cg.ir.CurrentFunction().Blocks, elseBlock)
	cg.currentBlock = elseBlock
	if ifStmt.ElseBody != nil {
		cg.generateBlock(ifStmt.ElseBody)
	}
	elseBlock.Terminator = &Instruction{
		Op:   "br",
		Args: []string{endLabel},
	}

	// End block
	endBlock := &BasicBlock{Label: endLabel}
	cg.ir.CurrentFunction().Blocks = append(cg.ir.CurrentFunction().Blocks, endBlock)
	cg.currentBlock = endBlock
}

func (cg *CodeGenerator) generateWhileStatement(whileStmt *parser.WhileStatement) {
	condLabel := cg.newLabel("while.cond")
	bodyLabel := cg.newLabel("while.body")
	endLabel := cg.newLabel("while.end")

	// Jump to condition
	cg.currentBlock.Terminator = &Instruction{
		Op:   "br",
		Args: []string{condLabel},
	}

	// Condition block
	condBlock := &BasicBlock{Label: condLabel}
	cg.ir.CurrentFunction().Blocks = append(cg.ir.CurrentFunction().Blocks, condBlock)
	cg.currentBlock = condBlock
	cond := cg.generateExpression(whileStmt.Condition)
	condBlock.Terminator = &Instruction{
		Op:   "br",
		Args: []string{cond, bodyLabel, endLabel},
	}

	// Body block
	bodyBlock := &BasicBlock{Label: bodyLabel}
	cg.ir.CurrentFunction().Blocks = append(cg.ir.CurrentFunction().Blocks, bodyBlock)
	cg.currentBlock = bodyBlock
	cg.generateBlock(whileStmt.Body)
	bodyBlock.Terminator = &Instruction{
		Op:   "br",
		Args: []string{condLabel},
	}

	// End block
	endBlock := &BasicBlock{Label: endLabel}
	cg.ir.CurrentFunction().Blocks = append(cg.ir.CurrentFunction().Blocks, endBlock)
	cg.currentBlock = endBlock
}

func (cg *CodeGenerator) generateForStatement(forStmt *parser.ForStatement) {
    // Removida a declaração não utilizada: initLabel
    condLabel := cg.newLabel("for.cond")
    bodyLabel := cg.newLabel("for.body")
    stepLabel := cg.newLabel("for.step")
    endLabel := cg.newLabel("for.end")

    // Initialization (executada no bloco atual)
    if forStmt.Init != nil {
        cg.generateStatement(forStmt.Init)
    }

    // Jump to condition
    cg.currentBlock.Terminator = &Instruction{
        Op:   "br",
        Args: []string{condLabel},
    }

    // Condition block
    condBlock := &BasicBlock{Label: condLabel}
    cg.ir.CurrentFunction().Blocks = append(cg.ir.CurrentFunction().Blocks, condBlock)
    cg.currentBlock = condBlock
    
    if forStmt.Condition != nil {
        cond := cg.generateExpression(forStmt.Condition)
        condBlock.Terminator = &Instruction{
            Op:   "br",
            Args: []string{cond, bodyLabel, endLabel},
        }
    } else {
        condBlock.Terminator = &Instruction{
            Op:   "br",
            Args: []string{bodyLabel},
        }
    }

    // Body block
    bodyBlock := &BasicBlock{Label: bodyLabel}
    cg.ir.CurrentFunction().Blocks = append(cg.ir.CurrentFunction().Blocks, bodyBlock)
    cg.currentBlock = bodyBlock
    cg.generateBlock(forStmt.Body)
    bodyBlock.Terminator = &Instruction{
        Op:   "br",
        Args: []string{stepLabel},
    }

    // Step block
    stepBlock := &BasicBlock{Label: stepLabel}
    cg.ir.CurrentFunction().Blocks = append(cg.ir.CurrentFunction().Blocks, stepBlock)
    cg.currentBlock = stepBlock
    if forStmt.Update != nil {
        cg.generateStatement(forStmt.Update)
    }
    stepBlock.Terminator = &Instruction{
        Op:   "br",
        Args: []string{condLabel},
    }

    // End block
    endBlock := &BasicBlock{Label: endLabel}
    cg.ir.CurrentFunction().Blocks = append(cg.ir.CurrentFunction().Blocks, endBlock)
    cg.currentBlock = endBlock
}
func (cg *CodeGenerator) generateReturnStatement(ret *parser.ReturnStatement) {
	if ret.Value != nil {
		val := cg.generateExpression(ret.Value)
		retType := cg.determineType(ret.Value)
		cg.currentBlock.Terminator = &Instruction{
			Op:   "ret",
			Type: retType,
			Args: []string{val},
		}
	} else {
		cg.currentBlock.Terminator = &Instruction{
			Op:   "ret",
			Type: VOID,
		}
	}
}

func (cg *CodeGenerator) generateBlock(block *parser.BlockStatement) {
	if block == nil {
		return
	}
	
	for _, stmt := range block.Statements {
		cg.generateStatement(stmt)
	}
}

func (cg *CodeGenerator) determineType(expr parser.Expression) Type {
	switch e := expr.(type) {
	case *parser.Number:
		if e.Value == float64(int(e.Value)) {
			return I32
		}
		return FLOAT
	case *parser.BooleanLiteral:
		return I1
	case *parser.Identifier:
		if info, exists := cg.symbolTable[e.Name]; exists {
			return info.Type
		}
	case *parser.UnaryExpression:
		return cg.determineType(e.Right)
	case *parser.BinaryExpression:
		leftType := cg.determineType(e.Left)
		rightType := cg.determineType(e.Right)
		if leftType == FLOAT || rightType == FLOAT {
			return FLOAT
		}
		return leftType
	}
	return I32
}

func (cg *CodeGenerator) llvmTypeFromParserType(t string) Type {
	switch t {
	case "int": return I32
	case "float": return FLOAT
	case "bool": return I1
	default: return I32
	}
}

func (cg *CodeGenerator) newTemp() string {
	temp := fmt.Sprintf("%%t%d", cg.tempCounter)
	cg.tempCounter++
	return temp
}

func (cg *CodeGenerator) newLabel(prefix string) string {
	label := fmt.Sprintf("%s%d", prefix, cg.labelCounter)
	cg.labelCounter++
	return label
}