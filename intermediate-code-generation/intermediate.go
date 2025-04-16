package intermediatecodegeneration

import "fmt"

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
    Functions   []*Function
    GlobalVars  []Instruction
    TempCounter int
    BlockCounter int
}

func NewIR() *IntermediateRep {
    return &IntermediateRep{
        Functions:  []*Function{{
            Name:       "main",
            ReturnType: I32,
            Blocks:     []*BasicBlock{{Label: "entry"}},
        }},
        TempCounter: 0,
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
    return label
}

func (ir *IntermediateRep) CurrentFunction() *Function {
    return ir.Functions[len(ir.Functions)-1]
}

func (ir *IntermediateRep) CurrentBlock() *BasicBlock {
    fn := ir.CurrentFunction()
    return fn.Blocks[len(fn.Blocks)-1]
}