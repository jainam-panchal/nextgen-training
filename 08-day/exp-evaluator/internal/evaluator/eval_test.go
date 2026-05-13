// internal/evaluator/eval_test.go
package evaluator_test

import (
	"errors"
	"exp-evaluator/internal/evaluator"
	"math"
	"strings"
	"testing"
)

func TestEvaluate_Integration(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		want    float64
		wantErr bool
		errIs   error
	}{
		// Basic arithmetic
		{"addition", "3 + 4", 7, false, nil},
		{"subtraction", "10 - 3", 7, false, nil},
		{"multiplication", "3 * 4", 12, false, nil},
		{"division", "8 / 2", 4, false, nil},

		// Operator precedence
		{"mul before add", "3 + 4 * 2", 11, false, nil},
		{"div before sub", "10 - 6 / 2", 7, false, nil},
		{"complex", "3 + 4 * 2 / (1 - 5) ^ 2", 3.5, false, nil},

		// Right-associative exponentiation
		{"right assoc pow", "2 ^ 3 ^ 2", 512, false, nil}, // 2^(3^2)=2^9=512

		// Parentheses override precedence
		{"parens", "(3 + 4) * 2", 14, false, nil},
		{"nested parens", "((2 + 3) * (4 - 1))", 15, false, nil},

		// Edge cases
		{"single number", "42", 42, false, nil},
		{"decimal", "3.14 * 2", 6.28, false, nil},
		{"negative result", "3 - 7", -4, false, nil},
		{"zero result", "5 - 5", 0, false, nil},

		// Error cases
		{"empty", "", 0, true, evaluator.ErrEmptyExpression},
		{"div by zero", "1 / 0", 0, true, evaluator.ErrDivisionByZero},
		{"missing operand", "+ 3", 0, true, nil}, // SyntaxError or MissingOperand
		{"unmatched open", "( 3 + 4", 0, true, nil},
		{"unmatched close", "3 + 4 )", 0, true, nil},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := evaluator.Evaluate(tc.expr)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("Evaluate(%q) = %v, want error", tc.expr, got)
				}
				if tc.errIs != nil && !errors.Is(err, tc.errIs) {
					t.Errorf("Evaluate(%q): errors.Is(%v) = false, got %v",
						tc.expr, tc.errIs, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Evaluate(%q) unexpected error: %v", tc.expr, err)
			}
			// Use tolerance for floating point comparison
			const epsilon = 1e-9
			if math.Abs(got-tc.want) > epsilon {
				t.Errorf("Evaluate(%q) = %v, want %v", tc.expr, got, tc.want)
			}
		})
	}
}

// Build: ((((... 1 + 1 ...)))) with 1000 levels
func TestEvaluate_DeepNesting(t *testing.T) {
	const depth = 1000
	expr := strings.Repeat("(", depth) + "1 + 1" + strings.Repeat(")", depth)

	result, err := evaluator.Evaluate(expr)
	if err != nil {
		t.Fatalf("Evaluate(depth=%d): unexpected error: %v", depth, err)
	}
	if result != 2 {
		t.Errorf("Evaluate(depth=%d) = %v, want 2", depth, result)
	}
}
