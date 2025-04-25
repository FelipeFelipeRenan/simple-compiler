package main

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"time"

	icg "simple-compiler/intermediate-code-generation"
	"simple-compiler/lexer"
	"simple-compiler/parser"
	"simple-compiler/token"
)

func main() {
	startingTime := time.Now()

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Uso: ./main <arquivo> [nome_output] [--run]")
		os.Exit(1)
	}

	fileName := os.Args[1]
	outputName := "output"
	shouldRun := false

	// Parse argumentos opcionais
	for _, arg := range os.Args[2:] {
		if arg == "--run" {
			shouldRun = true
		} else {
			outputName = arg
		}
	}

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
	generator := icg.NewCodeGenerator()
	intermediate := generator.GenerateFromAST(statements)

	if errs := generator.GetErrors(); len(errs) > 0 {
		fmt.Println("\nErros na geração de código:")
		for _, err := range errs {
			fmt.Printf("🔴 %s\n", err)
		}
		os.Exit(1)
	}

	generatedCode := intermediate.GenerateLLVM()
	fmt.Println("\n; Generated LLVM IR")
	fmt.Println(generatedCode)

	// 7. Salvar .ll
	tmpFile, err := os.CreateTemp("", outputName+"-*.ll")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao criar arquivo temporário LLVM IR: %v\n", err)
		os.Exit(1)
	}
	defer os.Remove(tmpFile.Name()) // Remove o arquivo temporário ao final
	
	if _, err := tmpFile.Write([]byte(generatedCode)); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao escrever no arquivo temporário: %v\n", err)
		os.Exit(1)
	}
	tmpFile.Close()
	
	llFile := tmpFile.Name()
	
	// 8. Gerar assembly com llc
	asmFile := outputName + ".s"
	cmdLLC := exec.Command("llc", llFile, "-o", asmFile)
	cmdLLC.Stdout = os.Stdout
	cmdLLC.Stderr = os.Stderr
	if err := cmdLLC.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao executar llc: %v\n", err)
		os.Exit(1)
	}

	// 9. Compilar com gcc -no-pie
	cmdGCC := exec.Command("gcc", "-no-pie", asmFile, "-o", outputName)
	cmdGCC.Stdout = os.Stdout
	cmdGCC.Stderr = os.Stderr
	if err := cmdGCC.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao compilar com gcc: %v\n", err)
		os.Exit(1)
	}

	// 10. Executar o binário gerado (se flag --run for usada)
	if shouldRun {
		fmt.Println("\n🔹 Saída do programa:")
		cmdExec := exec.Command("./" + outputName)
		out, _ := cmdExec.CombinedOutput()
		fmt.Print(string(out))

		// (Opcional) Erros de saída desativados como solicitado
		/*
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				fmt.Fprintf(os.Stderr, "⚠️ Código de saída do programa: %d\n", exitErr.ExitCode())
			} else {
				fmt.Fprintf(os.Stderr, "Erro ao executar o programa: %v\n", err)
			}
		}
		*/
	}

	elapsed := time.Since(startingTime)
	fmt.Printf("\n⏱️ Tempo de compilação total: %v\n", elapsed)
}

func sortErrorsByPosition(errors []parser.ParseError) {
	sort.Slice(errors, func(i, j int) bool {
		if errors[i].Line == errors[j].Line {
			return errors[i].Column < errors[j].Column
		}
		return errors[i].Line < errors[j].Line
	})
}