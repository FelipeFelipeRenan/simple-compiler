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
	errors       []string // Campo errors adicionado

}

type VariableInfo struct {
	Alloca string
	Type   Type
}

func NewCodeGenerator() *CodeGenerator {
	ir := NewIR()
	cg := &CodeGenerator{
		ir:           ir,
		symbolTable:  make(map[string]VariableInfo),
		tempCounter:  0,
		labelCounter: 0,
		errors:       make([]string, 0),
	}

	// Não cria bloco inicial automaticamente
	return cg
}

func (cg *CodeGenerator) GenerateFromAST(statements []parser.Statement) *IntermediateRep {
	// Primeiro processa declarações de função
	for _, stmt := range statements {
		if fnDecl, ok := stmt.(*parser.FunctionDeclaration); ok {
			cg.generateFunctionDecl(fnDecl)
		}
	}

	// Depois processa outras declarações
	for _, stmt := range statements {
		if _, ok := stmt.(*parser.FunctionDeclaration); !ok {
			if cg.currentBlock == nil {
				// Cria uma função main implícita se necessário
				if !cg.ir.hasFunction("main") {
					cg.generateImplicitMain()
				}
			}
			cg.generateStatement(stmt)
		}
	}

	return cg.ir
}

func (cg *CodeGenerator) generateStatement(stmt parser.Statement) {
	if cg.currentBlock == nil {
		cg.generateImplicitMain()
	}

	switch s := stmt.(type) {

	case *parser.VariableDeclaration:
		cg.generateVariableDecl(s, false)
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

func (cg *CodeGenerator) generateVariableDecl(decl *parser.VariableDeclaration, initializeOnly bool) {
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
			Args: []string{val, string(llvmType) + "*", alloca},
		})
	} else if initializeOnly {
		// Inicializa com valor padrão
		cg.currentBlock.Instructions = append(cg.currentBlock.Instructions, Instruction{
			Op:   "store",
			Type: llvmType,
			Args: []string{"0", string(llvmType) + "*", alloca},
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
		Args: []string{val, string(info.Type) + "*", info.Alloca},
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
		Args: []string{string(info.Type) + "*", info.Alloca},
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
	var predicate string // Novo: armazenar o predicado separadamente
	var cmpType Type = I1

	if leftType == FLOAT || rightType == FLOAT {
		op = "fcmp"
		switch expr.Operator {
		case "<":
			predicate = "olt"
		case ">":
			predicate = "ogt"
		case "<=":
			predicate = "ole"
		case ">=":
			predicate = "oge"
		case "==":
			predicate = "oeq"
		case "!=":
			predicate = "one"
		}
	} else {
		op = "icmp"
		switch expr.Operator {
		case "<":
			predicate = "slt"
		case ">":
			predicate = "sgt"
		case "<=":
			predicate = "sle"
		case ">=":
			predicate = "sge"
		case "==":
			predicate = "eq"
		case "!=":
			predicate = "ne"
		}
	}

	// Adiciona o predicado como primeiro argumento
	cg.currentBlock.Instructions = append(cg.currentBlock.Instructions, Instruction{
		Op:   op,
		Type: cmpType,
		Args: []string{predicate, string(leftType), left, right},
		Dest: temp,
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

	thenBlock := &BasicBlock{Label: thenLabel}
	cg.ir.CurrentFunction().Blocks = append(cg.ir.CurrentFunction().Blocks, thenBlock)
	cg.currentBlock = thenBlock
	cg.generateBlock(ifStmt.Body)
	thenBlock.Terminator = &Instruction{
		Op:   "br",
		Args: []string{endLabel},
	}

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

	endBlock := &BasicBlock{Label: endLabel}
	cg.ir.CurrentFunction().Blocks = append(cg.ir.CurrentFunction().Blocks, endBlock)
	cg.currentBlock = endBlock
}

func (cg *CodeGenerator) generateWhileStatement(whileStmt *parser.WhileStatement) {
	condLabel := cg.newLabel("while.cond")
	bodyLabel := cg.newLabel("while.body")
	endLabel := cg.newLabel("while.end")

	cg.currentBlock.Terminator = &Instruction{
		Op:   "br",
		Args: []string{condLabel},
	}

	condBlock := &BasicBlock{Label: condLabel}
	cg.ir.CurrentFunction().Blocks = append(cg.ir.CurrentFunction().Blocks, condBlock)
	cg.currentBlock = condBlock
	cond := cg.generateExpression(whileStmt.Condition)
	condBlock.Terminator = &Instruction{
		Op:   "br",
		Args: []string{cond, bodyLabel, endLabel},
	}

	bodyBlock := &BasicBlock{Label: bodyLabel}
	cg.ir.CurrentFunction().Blocks = append(cg.ir.CurrentFunction().Blocks, bodyBlock)
	cg.currentBlock = bodyBlock
	cg.generateBlock(whileStmt.Body)
	bodyBlock.Terminator = &Instruction{
		Op:   "br",
		Args: []string{condLabel},
	}

	endBlock := &BasicBlock{Label: endLabel}
	cg.ir.CurrentFunction().Blocks = append(cg.ir.CurrentFunction().Blocks, endBlock)
	cg.currentBlock = endBlock
}

func (cg *CodeGenerator) generateForStatement(forStmt *parser.ForStatement) {
	condLabel := cg.newLabel("for.cond")
	bodyLabel := cg.newLabel("for.body")
	stepLabel := cg.newLabel("for.step")
	endLabel := cg.newLabel("for.end")

	if forStmt.Init != nil {
		cg.generateStatement(forStmt.Init)
	}

	cg.currentBlock.Terminator = &Instruction{
		Op:   "br",
		Args: []string{condLabel},
	}

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

	bodyBlock := &BasicBlock{Label: bodyLabel}
	cg.ir.CurrentFunction().Blocks = append(cg.ir.CurrentFunction().Blocks, bodyBlock)
	cg.currentBlock = bodyBlock
	cg.generateBlock(forStmt.Body)
	bodyBlock.Terminator = &Instruction{
		Op:   "br",
		Args: []string{stepLabel},
	}

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
func (cg *CodeGenerator) generateFunctionDecl(decl *parser.FunctionDeclaration) {
	// Verifica se a função já existe
	if cg.ir.hasFunction(decl.Name) {
		cg.AddError(fmt.Sprintf("Função '%s' já declarada", decl.Name))
		return
	}

	// Converte tipo de retorno
	returnType := cg.llvmTypeFromParserType(decl.ReturnType)
	if decl.Name == "main" && decl.ReturnType == "void" {
		returnType = VOID
	}

	// Prepara parâmetros
	var params []Param
	for _, param := range decl.Parameters {
		params = append(params, Param{
			Name: param.Name,
			Type: cg.llvmTypeFromParserType(param.Type),
		})
	}

	// Cria função no IR
	fn := &Function{
		Name:       decl.Name,
		ReturnType: returnType,
		Params:     params,
		Blocks:     []*BasicBlock{{Label: "entry"}},
	}
	cg.ir.Functions = append(cg.ir.Functions, fn)
	cg.currentBlock = fn.Blocks[0]

	// Gera alocações para parâmetros
	for _, param := range params {
		alloca := cg.newTemp()
		cg.currentBlock.Instructions = append(cg.currentBlock.Instructions, Instruction{
			Op:   "alloca",
			Type: param.Type,
			Dest: alloca,
		})

		// Armazena o valor do parâmetro
		paramReg := "%" + param.Name
		cg.currentBlock.Instructions = append(cg.currentBlock.Instructions, Instruction{
			Op:   "store",
			Type: param.Type,
			Args: []string{paramReg, string(param.Type) + "*", alloca},
		})

		cg.symbolTable[param.Name] = VariableInfo{
			Alloca: alloca,
			Type:   param.Type,
		}
	}

	// Gera corpo da função
	block := &parser.BlockStatement{
		Statements: decl.Body,
	}
	cg.generateBlock(block)

	// Adiciona retorno padrão se necessário
	if cg.currentBlock.Terminator == nil {
		if decl.ReturnType == "void" {
			cg.currentBlock.Terminator = &Instruction{
				Op:   "ret",
				Type: VOID,
			}
		} else {
			cg.currentBlock.Terminator = &Instruction{
				Op:   "ret",
				Type: returnType,
				Args: []string{"0"},
			}
		}
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
	case "int":
		return I32
	case "float":
		return FLOAT
	case "bool":
		return I1
	default:
		return I32
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
func (cg *CodeGenerator) AddError(msg string) {
	cg.errors = append(cg.errors, msg)
}

// Método para obter erros
func (cg *CodeGenerator) GetErrors() []string {
	return cg.errors
}

func (cg *CodeGenerator) generateImplicitMain() {
	mainFn := &Function{
		Name:       "main",
		ReturnType: I32,
		Blocks:     []*BasicBlock{{Label: "entry"}},
	}
	cg.ir.Functions = append(cg.ir.Functions, mainFn)
	cg.currentBlock = mainFn.Blocks[0]
}
