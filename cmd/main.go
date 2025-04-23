package main

import (
	"fmt"
	"os"
	icg "simple-compiler/intermediate-code-generation"
	"simple-compiler/lexer"
	"simple-compiler/parser"
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

	// 5. Exibir AST
	if len(statements) > 0 {
		fmt.Println("\nAST gerada com sucesso:")
		for _, stmt := range statements {
			fmt.Println(stmt.String())
		}
	}

	// 6. Geração de código intermediário
    if len(p.Errors) == 0 {
        generator := icg.NewCodeGenerator()
        intermediate := generator.GenerateFromAST(statements)
        
        // Verifica erros usando o novo método GetErrors()
        if errs := generator.GetErrors(); len(errs) > 0 {
            fmt.Println("\nErros na geração de código:")
            for _, err := range errs {
                fmt.Printf("🔴 %s\n", err)
            }
            os.Exit(1)
        }

        fmt.Println("\n; Generated LLVM IR")
        fmt.Println(intermediate.GenerateLLVM())
    }

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
