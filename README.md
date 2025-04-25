# Simple Compiler

Este projeto é um compilador simples desenvolvido em Go, com suporte básico à análise léxica, análise sintática, geração de AST e geração de código intermediário em LLVM IR. O compilador é capaz de compilar uma linguagem imperativa básica e gerar executáveis reais via `llc` e `gcc`.

[![FelipeFelipeRenan/simple-compiler context](https://badge.forgithub.com/FelipeFelipeRenan/simple-compiler)](https://uithub.com/FelipeFelipeRenan/simple-compiler)

---

## 🧱 Funcionalidades

- Analisador léxico
- Parser com geração de AST
- Tipagem básica (`int`, `void`)
- Funções (`func`), chamadas e retorno
- Comandos de controle (`while`, `return`)
- Geração de código LLVM IR intermediário
- Integração com `llc` e `gcc` para gerar binários executáveis
- Suporte à função `print()` mapeada para `printf` do C

---

## 📦 Requisitos

- [Go](https://golang.org) 1.18 ou superior
- [LLVM](https://llvm.org) com `llc` instalado
- [GCC](https://gcc.gnu.org) para gerar o executável final

---

## 🚀 Como usar

### Compilar e rodar um programa:
```bash
go run cmd/main.go input.txt --run
```
### Compilar e gerar o executável sem executar
```bash
go run cmd/main.go input.txt meu_programa
```


## Exemplo de código-fonte
```go
func sum(int a, int b) int {
    return a + b
}

func main() void {
    int result = sum(2, -10)
    print(result)
}
```
