package main

import (
	"fmt"
	"os"
	icg "simple-compiler/intermediate-code-generation" // Descomente esta linha
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
		
		fmt.Println("\nCódigo Intermediário Gerado:")
		for _, instr := range intermediate.Instructions {
			switch instr.Op {
			case icg.ASSIGN:
				fmt.Printf("%s = %s\n", instr.Dest, instr.Arg1)
			case icg.LABEL:
				fmt.Printf("%s:\n", instr.Label)
			case icg.IF_FALSE:
				fmt.Printf("if_false %s goto %s\n", instr.Arg1, instr.Label)
			case icg.GOTO:
				fmt.Printf("goto %s\n", instr.Label)
			case icg.CALL:
				fmt.Printf("%s = call %s(%s)\n", instr.Dest, instr.Arg1, instr.Arg2)
			default:
				fmt.Printf("%s = %s %s %s\n", instr.Dest, instr.Arg1, instr.Op, instr.Arg2)
			}
		}
	}

	elapsed := time.Since(startingTime)
	fmt.Printf("\nTempo de compilação: %v\n", elapsed)
}

// ... (mantenha as funções filterErrors e sortErrorsByPosition como estão)
// Função para filtrar erros duplicados
func filterErrors(errors []parser.ParseError) []parser.ParseError {
	var filtered []parser.ParseError
	seen := make(map[string]bool)

	for _, err := range errors {
		key := fmt.Sprintf("%d:%d:%s", err.Line, err.Column, err.Message)
		if !seen[key] && err.Line > 0 { // Ignora erros sem linha definida
			seen[key] = true
			filtered = append(filtered, err)
		}
	}
	return filtered
}

// Função para ordenar erros por posição no código
func sortErrorsByPosition(errors []parser.ParseError) {
	sort.Slice(errors, func(i, j int) bool {
		if errors[i].Line == errors[j].Line {
			return errors[i].Column < errors[j].Column
		}
		return errors[i].Line < errors[j].Line
	})
}
