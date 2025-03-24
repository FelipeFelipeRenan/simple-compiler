package parser

import "fmt"

// Representa a tabela de simbolos

type SymbolTable struct {
	variables map[string]ValueType
}

// Criar uma nova tabela de simbolos
func NewSymbolTable() *SymbolTable{
	return &SymbolTable{variables: make(map[string]ValueType)}
}

// Atualiza ou adiciona um simbolo a tabela
func (st *SymbolTable) Set(name string, expr Expression) {
	switch v := expr.(type) {
	case *Number:
		st.variables[name] = v.Value
	case *Identifier:
		st.variables[name] = v.Value
	default:
		fmt.Println("Erro: Tipo inválido para atribuição!")
	}
}

// Retorna o valor associado a uma variavel
func (st *SymbolTable) Get(name string) (ValueType, bool){
	value, exists := st.variables[name]
	return value, exists
}

// Imprime a tabela de simbolos (DEBUG)
func(st *SymbolTable) Debug(){
	fmt.Println("----Tabela de simbolos----")
	for k, v := range st.variables {
		fmt.Printf("%s = %v\n", k,v)
	}
}