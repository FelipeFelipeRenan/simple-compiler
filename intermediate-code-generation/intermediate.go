package intermediatecodegeneration

import "fmt"

type Operation string

const(
	ASSIGN Operation = "="
	ADD Operation = "+"
	SUB Operation = "-"
	MULT Operation = "*"
	DIV Operation = "/"

	GOTO    Operation = "goto"
    IFLT    Operation = "if<"
    IFLE    Operation = "if<="
    IFGT    Operation = "if>"
    IFGE    Operation = "if>="
    IFEQ    Operation = "if=="
    IFNE    Operation = "if!="
    LABEL   Operation = "label"

	// definir outras operações
)

type Instruction struct {
	Op Operation
	Dest string // Destino: Pode ser uma variavel ou um temporario
	Arg1 string // Primeiro argumento da operação
	Arg2 string // Segundo argumento, em operações binarias
	Label string // Para instruções de fluxo de controle
}

type IntermediateRep struct {
	Instructions []Instruction
	TempCount int // contador para variavies temporarias 
}

func NewIR() *IntermediateRep{
	return &IntermediateRep{
		Instructions: make([]Instruction, 0),
		TempCount: 0,
	}
}

func (ir *IntermediateRep) NewTemp()string{
	temp := fmt.Sprintf("t%d", ir.TempCount)
	ir.TempCount++
	return temp
}