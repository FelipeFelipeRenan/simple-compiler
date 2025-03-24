package main

import (
	"fmt"
	"simple-compiler/lexer"
	"simple-compiler/parser"
)

func main() {
	source := "x = 1"

	// Tokenizar código-fonte
	tokens := lexer.Tokenize(source)

	// Criar parser e executar análise
	p := parser.New(tokens)
	astNode := p.ParseAssignment()

	// Exibir resultado
	fmt.Println(astNode.String()) // x = (5 + 3)
}
