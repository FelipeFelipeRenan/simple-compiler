package main

import (
	"fmt"
	"simple-compiler/lexer"
	"simple-compiler/token"
)

func main() {
	input := "x = 10 + 5 - 20 + ;"

	l := lexer.New(input)

	for {
		tok := l.NextToken()
		fmt.Printf("Token: %+v\n", tok)
		if tok.Type == token.EOF {
			break
		}
	}
}
