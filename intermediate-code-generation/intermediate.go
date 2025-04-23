package intermediatecodegeneration

import (
	"fmt"
	"strings"
)

type Type string

const (
	I32   Type = "i32"
	FLOAT Type = "float"
	I1    Type = "i1"
	VOID  Type = "void"
)

type Instruction struct {
	Op      string
	Type    Type
	Dest    string
	Args    []string
	Label   string
	Comment string
}

type BasicBlock struct {
	Label        string
	Instructions []Instruction
	Terminator   *Instruction
}

type Function struct {
	Name       string
	ReturnType Type
	Params     []Param
	Blocks     []*BasicBlock
}

type Param struct {
	Name string
	Type Type
}

type IntermediateRep struct {
	Functions    []*Function
	GlobalVars   []Instruction
	TempCounter  int
	BlockCounter int
}

func NewIR() *IntermediateRep {
	return &IntermediateRep{
		Functions: []*Function{{
			Name:       "main",
			ReturnType: I32,
			Blocks:     []*BasicBlock{{Label: "entry"}},
		}},
		TempCounter:  0,
		BlockCounter: 0,
	}
}

func (ir *IntermediateRep) NewTemp() string {
	temp := fmt.Sprintf("%%t%d", ir.TempCounter)
	ir.TempCounter++
	return temp
}

func (ir *IntermediateRep) NewLabel(prefix string) string {
	label := fmt.Sprintf("%s.%d", prefix, ir.BlockCounter)
	ir.BlockCounter++
	return "%" + label
}

func (ir *IntermediateRep) CurrentFunction() *Function {
	return ir.Functions[len(ir.Functions)-1]
}

func (ir *IntermediateRep) CurrentBlock() *BasicBlock {
	fn := ir.CurrentFunction()
	return fn.Blocks[len(fn.Blocks)-1]
}

func (ir *IntermediateRep) GenerateLLVM() string {
	var code strings.Builder

	// Cabeçalho do módulo
	code.WriteString("; ModuleID = 'simple-compiler'\n")
	code.WriteString("source_filename = \"simple-compiler\"\n\n")

	// Declaração de funções intrínsecas
	code.WriteString("declare i32 @printf(i8*, ...)\n\n")

	// Funções definidas pelo usuário
	for _, fn := range ir.Functions {
		// Assinatura da função
		code.WriteString(fmt.Sprintf("define %s @%s(", fn.ReturnType, fn.Name))

		// Parâmetros
		params := make([]string, len(fn.Params))
		for i, param := range fn.Params {
			params[i] = fmt.Sprintf("%s %%%s", param.Type, param.Name)
		}
		code.WriteString(strings.Join(params, ", "))
		code.WriteString(") {\n")

		// Corpo da função
		for _, block := range fn.Blocks {
			if block.Label != "" && block.Label != "entry" {
				code.WriteString(block.Label + ":\n")
			}

			// Instruções
			for _, inst := range block.Instructions {
				code.WriteString("  " + inst.Format() + "\n")
			}

			// Terminador
			if block.Terminator != nil {
				code.WriteString("  " + block.Terminator.Format() + "\n")
			}
		}
		code.WriteString("}\n\n")
	}

	// Função main padrão se não existir
	if !ir.hasFunction("main") {
		code.WriteString(`define i32 @main() {
entry:
  ret i32 0
}
`)
	}

	return code.String()
}

func (i Instruction) Format() string {
	switch i.Op {
	case "icmp", "fcmp":
		if len(i.Args) != 4 {
			return fmt.Sprintf("; ERROR: %s instruction with invalid arguments", i.Op)
		}
		// Formato: %dest = icmp <predicate> <type> <op1>, <op2>
		return fmt.Sprintf("%s = %s %s %s %s, %s",
			i.Dest, i.Op, i.Args[0], i.Args[1], i.Args[2], i.Args[3])
	case "load":
		return fmt.Sprintf("%s = %s %s, %s %s", i.Dest, i.Op, i.Type, i.Args[0], i.Args[1])
	case "store":
		return fmt.Sprintf("%s %s %s, %s %s", i.Op, i.Type, i.Args[0], i.Args[1], i.Args[2])
	case "br":
		if len(i.Args) == 1 {
			return fmt.Sprintf("%s label %%%s", i.Op, i.Args[0])
		}
		return fmt.Sprintf("%s i1 %s, label %%%s, label %%%s", i.Op, i.Args[0], i.Args[1], i.Args[2])
	case "ret":
		if i.Type == "void" {
			return "ret void"
		}
		return fmt.Sprintf("ret %s %s", i.Type, i.Args[0])
	default:
		if i.Dest != "" {
			return fmt.Sprintf("%s = %s %s %s", i.Dest, i.Op, i.Type, strings.Join(i.Args, ", "))
		}
		return fmt.Sprintf("%s %s %s", i.Op, i.Type, strings.Join(i.Args, ", "))
	}
}
func (ir *IntermediateRep) hasFunction(name string) bool {
	for _, fn := range ir.Functions {
		if fn.Name == name {
			return true
		}
	}
	return false
}
