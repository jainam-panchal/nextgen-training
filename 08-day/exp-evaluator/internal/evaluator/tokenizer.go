package evaluator

import (
	"fmt"
	"strconv"
	"strings"
)

func Tokenize(expr string) ([]Token, error) {
	if strings.TrimSpace(expr) == "" {
		return nil, ErrEmptyExpression
	}

	var tokens []Token
	i := 0

	for i < len(expr) {
		ch := rune(expr[i])

		// skip space
		if ch == ' ' || ch == '\n' || ch == '\t' {
			i++
			continue
		}

		// multi digit / decimal number
		if isDigit(ch) || ch == '.' {
			j := i
			dotCount := 0

			for j < len(expr) && (isDigit(rune(expr[j])) || expr[j] == '.') {
				if expr[j] == '.' {
					dotCount++
				}
				j++
			}

			if dotCount > 1 {
				return nil, &SyntaxError{
					Input:    expr,
					Position: i,
					Message:  fmt.Sprintf("invalid number %q: multiple decimal points", expr[i:j]),
				}
			}

			num, err := strconv.ParseFloat(expr[i:j], 64)
			if err != nil {
				return nil, &SyntaxError{
					Input:    expr,
					Position: i,
					Message:  fmt.Sprintf("cannnot parse number %q: %v", expr[i:j], err),
				}
			}

			tokens = append(tokens, Token{Type: TokenNumber, Value: num, Pos: i})

			i = j
			continue
		}

		// Operator
		if IsOperator(ch) {
			tokens = append(tokens, Token{
				Type: TokenOperator,
				Op:   ch,
				Pos:  i,
			})

			i++
			continue
		}

		// Parentheses
		switch ch {
		case '(':
			tokens = append(tokens, Token{Type: TokenLeftParen, Pos: i})
		case ')':
			tokens = append(tokens, Token{Type: TokenRightParen, Pos: i})
		default:
			return nil, &SyntaxError{
				Input:    "",
				Position: 0,
				Message:  fmt.Sprintf("unexpected character %q", ch),
			}
		}
		i++
	}

	if len(tokens) == 0 {
		return nil, ErrEmptyExpression
	}

	return tokens, nil
}

func isDigit(r rune) bool { return r >= '0' && r <= '9' }
