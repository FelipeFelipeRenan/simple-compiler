package main

import (
	"fmt"
	"os"
	code_generator "simple-compiler/intermediate-code-generation"
	"simple-compiler/lexer"
	"simple-compiler/parser"
	"simple-compiler/token"
	"sort"
	"time"
)

func main() {
	startingTime := time.Now()
	fileName := os.Args[1]

	// 1. Ler o arquivo fonte
	source, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao ler o arquivo '%s': %v\n", fileName, err)
		os.Exit(1)
	}

	// 2. Análise Léxica com proteção contra loops
	maxTokens := len(source) * 3 // Limite generoso para evitar loops infinitos
	l := lexer.New(string(source))
	var tokens []token.Token

	// Coleta tokens com proteção contra loops
	for i := 0; i < maxTokens; i++ {
		tok := l.NextToken()
		tokens = append(tokens, tok)

		// Proteção contra tokens repetidos que podem indicar loop
		if i > 0 && tokens[i].Type == tokens[i-1].Type &&
			tokens[i].Lexeme == tokens[i-1].Lexeme &&
			tokens[i].Type != token.EOF {
			fmt.Fprintf(os.Stderr, "Erro: token repetido '%s' na linha %d\n",
				tok.Lexeme, tok.Line)
			os.Exit(1)
		}

		if tok.Type == token.EOF {
			break
		}
	}

	if len(tokens) >= maxTokens {
		fmt.Fprintln(os.Stderr, "Erro: limite máximo de tokens atingido - possível loop infinito")
		os.Exit(1)
	}

	// 3. Análise Sintática
	p := parser.New(tokens)
	statements := p.Parse()

	// 4. Processamento de erros
	// Processamento de erros
	if len(p.Errors) > 0 {
		fmt.Println("\nErros encontrados:")

		// Filtra erros duplicados
		errorSet := make(map[string]parser.ParseError)
		for _, err := range p.Errors {
			key := fmt.Sprintf("%d:%d:%s", err.Line, err.Column, err.Message)
			if _, exists := errorSet[key]; !exists && err.Line > 0 {
				errorSet[key] = err
			}
		}

		// Converte para slice e ordena
		var uniqueErrors []parser.ParseError
		for _, err := range errorSet {
			uniqueErrors = append(uniqueErrors, err)
		}
		sortErrorsByPosition(uniqueErrors)

		// Exibe erros
		for _, err := range uniqueErrors {
			fmt.Printf("🔴 Linha %d:%d - %s\n", err.Line, err.Column, err.Message)
		}
		os.Exit(1)
	}

	// Exibe AST se não houver erros
	if len(statements) > 0 {
		fmt.Println("\nAST gerada com sucesso:")
		for _, stmt := range statements {
			fmt.Println(stmt.String())
		}
	} else {
		fmt.Println("Nenhuma declaração válida encontrada")
	}

	// 5. Exibição da AST (apenas se não houver erros)
	if len(statements) > 0 {
		fmt.Println("\nAST gerada com sucesso:")
		for _, stmt := range statements {
			if stmt != nil {
				fmt.Println(stmt.String())
			}
		}
	} else {
		fmt.Println("Nenhuma declaração válida encontrada no código fonte")
	}

	if len(p.Errors) == 0{

		generator := code_generator.NewCodeGenerator()
		intermediate := generator.GenerateFromAST(statements)

		fmt.Println("\nCódigo intermediario:")
		for _, instr := range intermediate.Instructions {
			if instr.Op == code_generator.ASSIGN{
				fmt.Printf("%s = %s\n", instr.Dest, instr.Arg1)
			} else{
				fmt.Printf("%s = %s %s %s\n", instr.Dest, instr.Arg1, instr.Op, instr.Arg2)
			}
			
		}
	}
elapsed := time.Since(startingTime)
fmt.Println("Tempo de compilação: ", elapsed)
}

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
