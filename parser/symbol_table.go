package parser

import "fmt"

// SymbolTable representa a tabela de símbolos
type SymbolTable struct {
	variables map[string]Expression // Agora armazenamos Expression em vez de ValueType
}

// NewSymbolTable cria uma nova tabela de símbolos
func NewSymbolTable() *SymbolTable {
	return &SymbolTable{variables: make(map[string]Expression)}
}

// Set atualiza ou adiciona um símbolo à tabela
func (st *SymbolTable) Set(name string, expr Expression) {
	st.variables[name] = expr // Agora aceita qualquer Expression (Number, Identifier, BinaryExpression)
}

// Get retorna o valor associado a uma variável
func (st *SymbolTable) Get(name string) (Expression, bool) {
	value, exists := st.variables[name]
	return value, exists
}

// Debug imprime a tabela de símbolos (DEBUG)
func (st *SymbolTable) Debug() {
	fmt.Println("---- Tabela de Símbolos ----")
	for k, v := range st.variables {
		fmt.Printf("%s = %v\n", k, v.String()) // Usando String() para formatar a saída
	}
}
