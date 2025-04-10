package intermediatecodegeneration

type CodeGenerator struct {
	ir *IntermediateRep
}

func NewCodeGenerator() *CodeGenerator{
	return &CodeGenerator{
		ir: NewIR(),
	}
}


