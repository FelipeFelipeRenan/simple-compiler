package parser

import "fmt"

type SymbolTable struct {
	scopes []map[string]interface{} // Pilha de escopos
}

// Criar nova tabela de símbolos
func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		scopes: []map[string]interface{}{{}}, // Começa com um escopo global
	}
}
// Entra em um novo escopo
func (st *SymbolTable) PushScope(){
	st.scopes = append(st.scopes, map[string]interface{}{})
}

// Sai do escopo atual
func (st *SymbolTable) PopScope(){
	if len(st.scopes) > 1{
		st.scopes = st.scopes[:len(st.scopes)-1]
	} else {
		fmt.Println("Erro: Tentativa de sair do escopo global")
	}
}

// Set atualiza ou adiciona um símbolo à tabela
func (st *SymbolTable) Set(name string, value interface{}) {
	st.scopes[len(st.scopes)-1][name] = value // Agora aceita qualquer Expression (Number, Identifier, BinaryExpression)
}

// Get retorna o valor associado a uma variável
func (st *SymbolTable) Get(name string) (interface{}, bool) {
	for i := len(st.scopes) - 1; i > 0; i-- {
		if val, exists := st.scopes[i][name]; exists{
			return val, true
		}
	}
	return nil, false
}
