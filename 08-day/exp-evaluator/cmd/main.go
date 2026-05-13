// cmd/main.go
package main

import (
	"bufio"
	"exp-evaluator/internal/evaluator"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Expression Evaluator — type 'quit' to exit")
	fmt.Println("Supported: +, -, *, /, ^ and ( )")
	fmt.Println()

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break // EOF
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "quit" || line == "exit" || line == "q" {
			fmt.Println("Goodbye.")
			break
		}

		result, err := evaluator.Evaluate(line)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}
		fmt.Printf("= %g\n", result)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "stdin: %v\n", err)
		os.Exit(1)
	}
}
