package main

import (
	"fmt"
	"os"
	"simple-compiler/lexer"
	"simple-compiler/parser"
	"simple-compiler/semantic"
	"simple-compiler/token"
)

func main() {
	fileName := "input.txt"

	// 1. Ler o arquivo fonte
	source, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao ler o arquivo '%s': %v\n", fileName, err)
		os.Exit(1)
	}

	// 2. Análise Léxica
	l := lexer.New(string(source))
	var tokens []token.Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == token.EOF {
			break
		}
	}

	// 3. Análise Sintática
	p := parser.New(tokens)

	statements := p.Parse()

	// 4. Verificar erros de parsing
	if len(statements) == 0 {
		fmt.Println("Nenhuma declaração válida foi encontrada no código-fonte.")
		return
	}

	// 5. Análise Semântica
	analyzer := semantic.New(statements)
	semanticErrors := analyzer.Analyze()

	// 6. Exibir erros semânticos
	if len(semanticErrors) > 0 {
		fmt.Println("\nErros semânticos encontrados:")
		for _, err := range semanticErrors {
			fmt.Println("🔴", err)
		}
		os.Exit(1)
	}

	// 7. Exibir AST (apenas se não houver erros)
	fmt.Println("\nAST gerada:")
	for _, stmt := range statements {
		if stmt != nil {
			fmt.Println(stmt.String())
		}
	}
}
