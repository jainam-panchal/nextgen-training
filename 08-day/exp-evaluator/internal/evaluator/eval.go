package evaluator

import "fmt"

func Evaluate(expr string) (float64, error) {
	tokens, err := Tokenize(expr)
	if err != nil {
		return 0, fmt.Errorf("tokenize: %w", err)
	}

	postfix, err := InfixToPostfix(tokens)
	if err != nil {
		return 0, fmt.Errorf("parse: %w", err)
	}

	result, err := EvaluatePostFix(postfix)
	if err != nil {
		return 0, fmt.Errorf("evaluate: %w", err)
	}

	return result, nil
}
