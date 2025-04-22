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

	for _, fn := range ir.Functions {
		// Function header
		code.WriteString(fmt.Sprintf("define %s @%s() {\n", fn.ReturnType, fn.Name))

		// Basic blocks
		for _, block := range fn.Blocks {
			// Block label
			if block.Label != "" {
				code.WriteString(block.Label + ":\n")
			}

			// Instructions
			for _, inst := range block.Instructions {
				if inst.Op == "icmp" {
					fmt.Printf("DEBUG ICMP INSTRUCTION: %+v\n", inst)
				}
				code.WriteString("  " + inst.Format() + "\n")
			}

			// Terminator
			if block.Terminator != nil {
				code.WriteString("  " + block.Terminator.Format() + "\n")
			}
		}

		code.WriteString("}\n")
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