package parser

import "fmt"

// Definição de tipos
const (
	TypeInt   = 0
	TypeFloat = 1
)

// ValueType representa um valor com seu tipo
type ValueType struct {
	Value float64
	Type  int
}

func (v ValueType) String() string {
	return fmt.Sprintf("%v", v.Value) // Suporta tanto int quanto float
}
