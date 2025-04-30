# Simple Compiler

Este é um compilador simples desenvolvido em Go que traduz uma linguagem imperativa básica para LLVM IR, permitindo gerar executáveis reais utilizando `llc` e `gcc`. O projeto tem fins didáticos e demonstra de forma modular os principais estágios de um compilador: análise léxica, análise sintática, geração de AST e código intermediário.

---

## 🧠 Funcionalidades

- ✅ Análise léxica com geração de tokens
- ✅ Análise sintática e construção de AST
- ✅ Geração de código LLVM IR
- ✅ Integração com `llc` para gerar assembly
- ✅ Compilação final com `gcc -no-pie`
- ✅ Execução opcional do binário
- ✅ Suporte a `int`, `void`, `func`, `while`, `return`, `print`

---

## 📁 Estrutura do projeto

```
simple-compiler/
├── cmd/
│   └── main.go                      # Entrada principal do compilador
├── lexer/                           # Analisador léxico
├── parser/                          # Parser e AST
├── intermediate-code-generation/   # Gerador de LLVM IR
├── token/                           # Definição dos tokens
└── input.txt                        # Código de entrada exemplo
```

---

## ⚙️ Requisitos

- [Go](https://golang.org/dl/) 1.18 ou superior
- [LLVM](https://llvm.org/) com `llc` disponível no PATH
- [GCC](https://gcc.gnu.org/) com suporte a `-no-pie`

---

## 🚀 Como compilar e executar

### Compilar e gerar executável (sem rodar):
```bash
go run cmd/main.go input.txt meu_programa
```

### Compilar e executar:
```bash
go run cmd/main.go input.txt meu_programa --run
```

### Somente compilar (gera binário como `output` por padrão):
```bash
go run cmd/main.go input.txt
```

---
### Rodar utilizando a build do compilador
```bash
./gopher <mesmas opções acima>
```
## 💻 Exemplo de código fonte (`input.txt`)

```c
func sum(int a, int b) int {
    return a + b
}

func main() void {
    int result = sum(2, -10)
    print(result)
}
```

---

## 📦 Saída

O compilador irá:

1. Gerar arquivos `.ll` e `.s` temporários
2. Compilar o código em um binário com nome definido (ou `output` se não especificado)
3. Opcionalmente, executar o binário se `--run` for fornecido

---

## 🛠️ Melhorias futuras

- [ ] Suporte a estruturas condicionais (`if`, `else`)
- [ ] Análise semântica completa
- [ ] Tipos adicionais (bool, float, string)
- [ ] Suporte a escopos e funções aninhadas
- [ ] Otimizações no LLVM IR

---

