package intermediatecodegeneration

import (
	"fmt"
	"simple-compiler/parser"
	"strconv"
	"strings"
)

type CodeGenerator struct {
	ir           *IntermediateRep
	symbolTable  map[string]VariableInfo
	currentBlock *BasicBlock
	tempCounter  int
	labelCounter int
	stringCounter int
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
		stringCounter: 0,
		errors:       make([]string, 0),
	}

	// Não cria bloco inicial automaticamente
	return cg
}

func (cg *CodeGenerator) GenerateFromAST(statements []parser.Statement) *IntermediateRep {
	// Primeiro processa declarações de função
	cg.addPrintfSupport()

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
		if call, ok := s.Expression.(*parser.CallExpression); ok && call.FunctionName == "print" {
			cg.generatePrintCall(call)
		} else {
			cg.generateExpression(s.Expression)
		}
	}
}

func (cg *CodeGenerator) generateVariableDecl(decl *parser.VariableDeclaration, initializeOnly bool) {
    llvmType := cg.llvmTypeFromParserType(decl.Type)
    alloca := cg.newTemp()

    // Para strings, usamos i8* no lugar do tipo original
    storageType := llvmType
    if decl.Type == "string" {
        storageType = "i8*"
    }

    cg.currentBlock.Instructions = append(cg.currentBlock.Instructions, Instruction{
        Op:   "alloca",
        Type: storageType,
        Dest: alloca,
    })

    cg.symbolTable[decl.Name] = VariableInfo{
        Alloca: alloca,
        Type:   llvmType, // Mantemos o tipo original na tabela de símbolos
    }

    if decl.Value != nil {
        val := cg.generateExpression(decl.Value)
        cg.currentBlock.Instructions = append(cg.currentBlock.Instructions, Instruction{
            Op:   "store",
            Type: storageType,
            Args: []string{val, string(storageType) + "*", alloca},
        })
    } else if initializeOnly {
        // Inicializa com valor padrão
        defaultVal := "0"
        if decl.Type == "string" {
            defaultVal = "null"
        }
        cg.currentBlock.Instructions = append(cg.currentBlock.Instructions, Instruction{
            Op:   "store",
            Type: storageType,
            Args: []string{defaultVal, string(storageType) + "*", alloca},
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
	case *parser.StringLiteral:
        return cg.generateStringLiteral(e)
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
    
    // Tratamento especial para strings
    if info.Type == "string" {
        cg.currentBlock.Instructions = append(cg.currentBlock.Instructions, Instruction{
            Op:   "load",
            Type: "i8*",
            Args: []string{"i8**", info.Alloca},
            Dest: temp,
        })
    } else {
        cg.currentBlock.Instructions = append(cg.currentBlock.Instructions, Instruction{
            Op:   "load",
            Type: info.Type,
            Args: []string{string(info.Type) + "*", info.Alloca},
            Dest: temp,
        })
    }
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
	if call.FunctionName == "print" {
		if len(call.Arguments) != 1 {
			cg.AddError("print requer exatamente 1 argumento")
			return "0"
		}

		arg := cg.generateExpression(call.Arguments[0])
		argType := cg.determineType(call.Arguments[0])

		fmtStr := cg.newTemp()
		formatStr := "@.str" // padrão para int
		if argType == FLOAT {
			formatStr = "@.strf"
		}

		cg.currentBlock.Instructions = append(cg.currentBlock.Instructions, Instruction{
			Op:   "getelementptr",
			Dest: fmtStr,
			Args: []string{fmt.Sprintf("[4 x i8], [4 x i8]* %s, i32 0, i32 0", formatStr)},
		})

		cg.currentBlock.Instructions = append(cg.currentBlock.Instructions, Instruction{
			Op:   "call",
			Type: I32,
			Args: []string{fmt.Sprintf("i32 (i8*, ...) @printf(i8* %s, %s %s)", fmtStr, argType, arg)},
		})

		return "0"
	}
	// Restante da implementação original...
	temp := cg.newTemp()
	args := make([]string, len(call.Arguments))

	// Processa os argumentos
	for i, arg := range call.Arguments {
		args[i] = cg.generateExpression(arg)
	}

	// Obtém o tipo de retorno da função
	returnType := cg.getFunctionReturnType(call.FunctionName)

	// Formata os argumentos com seus tipos
	typedArgs := make([]string, len(args))
	for i, arg := range args {
		argType := cg.determineType(call.Arguments[i])
		typedArgs[i] = fmt.Sprintf("%s %s", argType, arg)
	}

	cg.currentBlock.Instructions = append(cg.currentBlock.Instructions, Instruction{
		Op:   "call",
		Type: returnType,
		Dest: temp,
		Args: []string{
			fmt.Sprintf("%s @%s(%s)", returnType, call.FunctionName, strings.Join(typedArgs, ", ")),
		},
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
	case *parser.StringLiteral:
		return I8
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
	case "string":
        return I8
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

func (cg *CodeGenerator) getFunctionReturnType(funcName string) Type {
	// Verifica nas funções geradas
	for _, fn := range cg.ir.Functions {
		if fn.Name == funcName {
			return fn.ReturnType
		}
	}

	// Funções padrão (como print, etc) em breve, maybeee
	switch funcName {
	case "":
		return I32
	// Adicione outros casos conforme necessário
	default:
		return I32 // Padrão para funções desconhecidas
	}
}

func (cg *CodeGenerator) addPrintfSupport() {
	cg.ir.GlobalVars = append(cg.ir.GlobalVars, Instruction{
		Op:   "declare",
		Args: []string{"i32 @printf(i8*, ...)"},
	})

	// Formato para inteiros
	cg.ir.GlobalVars = append(cg.ir.GlobalVars, Instruction{
		Op:   "@.str.int",
		Args: []string{"= private unnamed_addr constant [4 x i8] c\"%d\\0A\\00\", align 1"},
	})

	// Formato para floats
	cg.ir.GlobalVars = append(cg.ir.GlobalVars, Instruction{
		Op:   "@.str.float",
		Args: []string{"= private unnamed_addr constant [4 x i8] c\"%f\\0A\\00\", align 1"},
	})

	// Formato para strings
	cg.ir.GlobalVars = append(cg.ir.GlobalVars, Instruction{
		Op:   "@.str.str",
		Args: []string{"= private unnamed_addr constant [4 x i8] c\"%s\\0A\\00\", align 1"},
	})
}
func (cg *CodeGenerator) generatePrintCall(call *parser.CallExpression) {
    if len(call.Arguments) != 1 {
        cg.AddError("print requer exatamente 1 argumento")
        return
    }
    
    arg := cg.generateExpression(call.Arguments[0])
    argType := cg.determineType(call.Arguments[0])
    
    var formatStr string
    var argTypeStr string
    
    switch argType {
    case I32:
        formatStr = "@.str.int"
        argTypeStr = "i32"
    case FLOAT:
        formatStr = "@.str.float"
        argTypeStr = "float"
    case "i8*":
        formatStr = "@.str.str"
        argTypeStr = "i8*"
    default:
        cg.AddError(fmt.Sprintf("Tipo não suportado para print: %s", argType))
        return
    }
    
    fmtPtr := cg.newTemp()
    cg.currentBlock.Instructions = append(cg.currentBlock.Instructions, Instruction{
        Op:   "getelementptr",
        Dest: fmtPtr,
        Args: []string{fmt.Sprintf("[4 x i8], [4 x i8]* %s, i32 0, i32 0", formatStr)},
    })
    
    // Adiciona a chamada ao printf com uma variável temporária para o resultado
    callResult := cg.newTemp()
    cg.currentBlock.Instructions = append(cg.currentBlock.Instructions, Instruction{
        Op:   "call",
        Dest: callResult,
        Type: I32,
        Args: []string{fmt.Sprintf("i32 (i8*, ...) @printf(i8* %s, %s %s)", fmtPtr, argTypeStr, arg)},
    })
}

func (cg *CodeGenerator) generateStringLiteral(str *parser.StringLiteral) string {
    strName := fmt.Sprintf("@.str.%d", cg.stringCounter)
    cg.stringCounter++
    
    strValue := str.Value + "\\00"
    strLen := len(str.Value) + 1
    
    cg.ir.GlobalVars = append(cg.ir.GlobalVars, Instruction{
        Op:    strName,
        Args:  []string{fmt.Sprintf("= private unnamed_addr constant [%d x i8] c\"%s\", align 1", strLen, strValue)},
    })
    
    temp := cg.newTemp()
    cg.currentBlock.Instructions = append(cg.currentBlock.Instructions, Instruction{
        Op:   "getelementptr",
        Dest: temp,
        Args: []string{fmt.Sprintf("[%d x i8], [%d x i8]* %s, i32 0, i32 0", strLen, strLen, strName)},
    })
    
    return temp
}