package parser


// tipo suportado pelo compilador
type Type int

const (
	TypeInt Type = iota
	TypeFloat
	TypeString
	TypeBool
	TypeUnknown // Tipo nao conhecido, gerando um erro
)

// ValueType representa um valor e seu tipo
type ValueType struct {
	Value interface{}
	Type  Type
}
