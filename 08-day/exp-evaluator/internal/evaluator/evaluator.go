package evaluator

import (
	"fmt"
	"math"

	"exp-evaluator/internal/stack"
)

func EvaluatePostFix(tokens []Token) (result float64, err error) {
	// must recover from any expected panic
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("internal evaluator error: %v", r)
			result = 0
		}
	}()

	operands := stack.NewStack[float64](len(tokens))

	for _, token := range tokens {
		switch token.Type {

		case TokenNumber:
			operands.Push(token.Value)

		case TokenOperator:
			if operands.Size() < 2 {
				return 0, fmt.Errorf("operator %q at position %d: %w", token.Op, token.Pos, ErrMissingOperand)
			}

			right := operands.MustPop()
			left := operands.MustPop()

			var opResult float64
			switch token.Op {
			case '+':
				opResult = left + right
			case '-':
				opResult = left - right
			case '*':
				opResult = left * right
			case '/':
				if right == 0 {
					return 0, fmt.Errorf("division at position %d: %w", token.Pos, ErrDivisionByZero)
				}
				opResult = left / right
			case '^':
				opResult = math.Pow(left, right)
			default:
				return 0, fmt.Errorf("unknown operator %q", token.Op)
			}
			operands.Push(opResult)
		}
	}

	if operands.Size() != 1 {
		return 0, fmt.Errorf("expression has %d leftover values: %w",
			operands.Size(), ErrMissingOperator)
	}

	return operands.MustPop(), nil
}
