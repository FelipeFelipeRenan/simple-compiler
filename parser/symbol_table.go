// parser/symbol_table.go
package parser

import (
	"fmt"
)

// Tipo para categorias de símbolos
type SymbolCategory int

const (
	Variable SymbolCategory = iota
	Function
	Constant
)

// Tipo para informações do símbolo
type SymbolInfo struct {
	Name      string
	Category  SymbolCategory
	Type      string      // Tipo do símbolo (int, float, etc)
	Value     interface{} // Valor atual (opcional)
	DefinedAt int         // Linha onde foi definido (para mensagens de erro)
}

// Tabela de símbolos com escopos aninhados
type SymbolTable struct {
	scopes []map[string]SymbolInfo
}

// Cria nova tabela de símbolos com escopo global
func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		scopes: []map[string]SymbolInfo{
			make(map[string]SymbolInfo), // Escopo global
		},
	}
}

// Entra em um novo escopo
func (st *SymbolTable) PushScope() {
	st.scopes = append(st.scopes, make(map[string]SymbolInfo))
}

// Sai do escopo atual
func (st *SymbolTable) PopScope() {
	if len(st.scopes) <= 1 {
		panic("cannot pop global scope")
	}
	st.scopes = st.scopes[:len(st.scopes)-1]
}

// Declara um novo símbolo no escopo atual
func (st *SymbolTable) Declare(name string, info SymbolInfo) error {
    currentScope := st.scopes[len(st.scopes)-1]
    
    if _, exists := currentScope[name]; exists {
        return fmt.Errorf("symbol '%s' already declared in this scope", name)
    }
    
    currentScope[name] = info
    return nil
}

func (st *SymbolTable) Resolve(name string) (SymbolInfo, bool) {
    for i := len(st.scopes) - 1; i >= 0; i-- {
        if info, exists := st.scopes[i][name]; exists {
            return info, true
        }
    }
    return SymbolInfo{}, false
}

// Atualiza o valor de um símbolo existente
func (st *SymbolTable) Update(name string, value interface{}) error {
	for i := len(st.scopes) - 1; i >= 0; i-- {
		if info, exists := st.scopes[i][name]; exists {
			// Verifica tipo antes de atualizar
			if info.Type == "int" {
				if _, ok := value.(int); !ok {
					return fmt.Errorf("type mismatch: expected int for %s", name)
				}
			}
			// Atualiza apenas o valor, mantendo outras informações
			info.Value = value
			st.scopes[i][name] = info
			return nil
		}
	}
	return fmt.Errorf("symbol '%s' not declared", name)
}

// Verifica se um símbolo existe no escopo atual (sem procurar nos pais)
func (st *SymbolTable) ExistsInCurrentScope(name string) bool {
	_, exists := st.scopes[len(st.scopes)-1][name]
	return exists
}