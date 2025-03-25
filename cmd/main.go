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

	// Ler o arquivo fonte
	source, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao ler o arquivo '%s': %v\n", fileName, err)
		os.Exit(1)
	}

	// Criar lexer
	l := lexer.New(string(source))

	// Coletar tokens
	var tokens []token.Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)

		if tok.Type == token.EOF {
			break
		}
	}

	// Criar parser e processar declarações
	p := parser.New(tokens)
	statements := p.Parse()

	// Verificar se o parsing resultou em erros
	if len(statements) == 0 {
		fmt.Println("Nenhuma declaração válida foi encontrada no código-fonte.")
		return
	}

	// Exibir AST
	fmt.Println("AST gerada:")
	for _, stmt := range statements {
		if stmt != nil {
			fmt.Println(stmt.String()) // Certifique-se que os statements implementam o método String()
		} else {
			fmt.Println("Erro: declaração inválida encontrada.")
		}
	}
}
