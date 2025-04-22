package main

import (
	"fmt"
	"os"
	icg "simple-compiler/intermediate-code-generation"
	"simple-compiler/lexer"
	"simple-compiler/parser"
	"simple-compiler/semantic"
	"simple-compiler/token"
	"sort"
	"time"
)

func main() {
	startingTime := time.Now()
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Uso: go run cmd/main.go <arquivo>")
		os.Exit(1)
	}
	fileName := os.Args[1]

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

	// Adicione isso temporariamente no cmd/main.go após a análise léxica
	fmt.Println("\nTokens gerados:")
	for _, tok := range tokens {
		fmt.Printf("Type: %-10s Lexeme: %-10s Line: %d Column: %d\n",
			tok.Type, tok.Lexeme, tok.Line, tok.Column)
	}
	// 3. Análise Sintática
	p := parser.New(tokens)
	statements := p.Parse()

	// 4. Processamento de erros
	if len(p.Errors) > 0 {
		fmt.Println("\nErros encontrados:")
		sortErrorsByPosition(p.Errors)
		for _, err := range p.Errors {
			fmt.Printf("🔴 Linha %d:%d - %s\n", err.Line, err.Column, err.Message)
		}
		os.Exit(1)
	}

	// 5. Análise Semântica
	analyzer := semantic.New(statements)
	semanticErrors := analyzer.Analyze()
	if len(semanticErrors) > 0 {
		fmt.Println("\nErros semânticos encontrados:")
		for _, err := range semanticErrors {
			fmt.Printf("🔴 Linha %d - %s\n", err.Line, err.Message)
		}
		os.Exit(1)
	}

	// 6. Geração de Código Intermediário
	generator := icg.NewCodeGenerator()
	intermediate := generator.GenerateFromAST(statements)

	fmt.Println("\n; Generated LLVM IR")
	fmt.Println(intermediate.GenerateLLVM())

	elapsed := time.Since(startingTime)
	fmt.Printf("\nTempo de compilação: %v\n", elapsed)
}
func sortErrorsByPosition(errors []parser.ParseError) {
	sort.Slice(errors, func(i, j int) bool {
		if errors[i].Line == errors[j].Line {
			return errors[i].Column < errors[j].Column
		}
		return errors[i].Line < errors[j].Line
	})
}
