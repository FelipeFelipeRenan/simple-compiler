package intermediatecodegeneration

import "fmt"

type Operation string

const (
	ASSIGN Operation = "="
	ADD    Operation = "+"
	SUB    Operation = "-"
	MULT   Operation = "*"
	DIV    Operation = "/"

	GOTO     Operation = "goto"
	IFLT     Operation = "if<"
	IFLE     Operation = "if<="
	IFGT     Operation = "if>"
	IFGE     Operation = "if>="
	IFEQ     Operation = "if=="
	IFNE     Operation = "if!="
	IF_FALSE Operation = "if_false"
	LABEL    Operation = "label"
	RETURN   Operation = "return"
	CALL     Operation = "call"
	// Operadores unários
	NEG Operation = "neg" // Para -
	NOT Operation = "not" // Para !

	// definir outras operações
)

type Instruction struct {
    Op     Operation
    Dest   string   // Para atribuições
    Arg1   string   // Primeiro operando
    Arg2   string   // Segundo operando ou lista de args
    Label  string   // Para controle de fluxo
}
type IntermediateRep struct {
	Instructions []Instruction
	TempCount    int // contador para variavies temporarias
}

func NewIR() *IntermediateRep {
	return &IntermediateRep{
		Instructions: make([]Instruction, 0),
		TempCount:    0,
	}
}

func (ir *IntermediateRep) NewTemp() string {
	temp := fmt.Sprintf("t%d", ir.TempCount)
	ir.TempCount++
	return temp
}
