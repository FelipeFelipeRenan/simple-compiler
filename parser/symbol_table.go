package parser

import "fmt"

// Representa a tabela de simbolos

type SymbolTable struct {
	variables map[string]interface{}
}

// Criar uma nova tabela de simbolos
func NewSymbolTable() *SymbolTable{
	return &SymbolTable{variables: make(map[string]interface{})}
}

// Atualiza ou adiciona um simbolo a tabela
func (st *SymbolTable) Set(name string, value interface{}){
	st.variables[name] = value
} 

// Retorna o valor associado a uma variavel
func (st *SymbolTable) Get(name string) (interface{}, bool){
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