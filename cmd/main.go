package main

import (
	"fmt"
	"simple-compiler/lexer"
	"simple-compiler/parser"
)

func main() {
	source := "x = 5 + 3 * 2 - 4 / 2"


	// Tokenizar código-fonte
	tokens := lexer.Tokenize(source)

	// Criar parser e executar análise
	p := parser.New(tokens)
	astNode := p.ParseAssignment()

	// Exibir resultado
	fmt.Println("Atribuição:", astNode.String()) // Agora exibe apenas o número corretamente
}
