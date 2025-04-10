package intermediatecodegeneration

type Operation string

const(
	ASSIGN Operation = "="
	ADD Operation = "+"
	SUB Operation = "-"
	MULT Operation = "*"
	DIV Operation = "/"

	// definir outras operações
)

type Instruction struct {
	Op Operation
	Dest string // Destino: Pode ser uma variavel ou um temporario
	Arg1 string // Primeiro argumento da operação
	Arg2 string // Segundo argumento, em operações binarias
	Label string // Para instruções de fluxo de controle
}