package main

import (
	"fmt"
	"os"
	"simple-compiler/lexer"
	"simple-compiler/parser"
	"simple-compiler/token"
)

func main() {
	fileName := "input.txt"
	source, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Println("Erro ao ler arquivo:", err)
		os.Exit(1)
	}

	// Criar lexer
	l := lexer.New(string(source))

	// Coletar tokens
	var tokens []token.Token
	for {
		tok := l.NextToken()
		if tok.Type == token.EOF {
			break
		}
		tokens = append(tokens, tok)
	}

	// Criar parser e processar declarações
	p := parser.New(tokens)
	statements := p.Parse()

	// Exibir AST
	for _, stmt := range statements {
		fmt.Println(stmt.String())
	}
}
