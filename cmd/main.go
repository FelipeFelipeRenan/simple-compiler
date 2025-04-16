package main

import (
	"fmt"
	"os"
	icg "simple-compiler/intermediate-code-generation" // Descomente esta linha
	"simple-compiler/lexer"
	"simple-compiler/parser"
	"simple-compiler/token"
	"sort"
	"strings"
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
// Modifique a parte da geração de código no main():
if len(p.Errors) == 0 {
    generator := icg.NewCodeGenerator()
    intermediate := generator.GenerateFromAST(statements)
    printLLVMIR(intermediate)
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

// Adicione esta função para imprimir a IR
func printLLVMIR(ir *icg.IntermediateRep) {
    fmt.Println("\n; Generated LLVM IR")
    
    for _, fn := range ir.Functions {
        // Print function header
        params := make([]string, len(fn.Params))
        for i, p := range fn.Params {
            params[i] = fmt.Sprintf("%s %s", p.Type, p.Name)
        }
        
        fmt.Printf("\ndefine %s @%s(%s) {\n", 
            fn.ReturnType, fn.Name, strings.Join(params, ", "))
        
        // Print basic blocks
        for _, block := range fn.Blocks {
            if block.Label != "" {
                fmt.Printf("%s:\n", block.Label)
            }
            
            // Print instructions
            for _, inst := range block.Instructions {
                if inst.Dest != "" {
                    fmt.Printf("  %s = ", inst.Dest)
                } else {
                    fmt.Printf("  ")
                }
                
                fmt.Printf("%s %s", inst.Op, inst.Type)
                
                if len(inst.Args) > 0 {
                    fmt.Printf(" %s", strings.Join(inst.Args, ", "))
                }
                
                if inst.Comment != "" {
                    fmt.Printf(" ; %s", inst.Comment)
                }
                
                fmt.Println()
            }
            
            // Print terminator
            if block.Terminator != nil {
                fmt.Printf("  %s %s", block.Terminator.Op, block.Terminator.Type)
                if len(block.Terminator.Args) > 0 {
                    fmt.Printf(" %s", strings.Join(block.Terminator.Args, ", "))
                }
                fmt.Println()
            }
        }
        
        fmt.Println("}")
    }
}

