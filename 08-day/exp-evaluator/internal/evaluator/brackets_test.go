package evaluator_test

import (
	"errors"
	"exp-evaluator/internal/evaluator"
	"testing"
)

func TestBracketMatch(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		wantErr bool
	}{
		{"empty", "", false},
		{"no brackets", "3 + 4 * 2", false},
		{"simple parens", "(3 + 4)", false},
		{"nested mixed", "({[3 + (4*2)]})", false},

		{"unmatched open paren", "(3 + 4", true},
		{"unmatched close paren", "3 + 4)", true},
		{"mismatched pair", "(3 + 4]", true},
		{"wrong order", ")(3+4)", true},
		{"unmatched open brace", "{[3+4]", true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := evaluator.BracketMatch(tc.expr)
			if tc.wantErr && err == nil {
				t.Fatalf("BracketMatch(%q): expected error, got nil", tc.expr)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("BracketMatch(%q): unexpected error: %v", tc.expr, err)
			}

			// Optional: ensure mismatch errors are typed.
			if tc.wantErr && err != nil {
				if _, ok := errors.AsType[*evaluator.MismatchError](err); !ok {
					t.Fatalf("BracketMatch(%q): expected MismatchError, got %T", tc.expr, err)
				}
			}
		})
	}
}
