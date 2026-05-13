package evaluator

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrEmptyExpression = errors.New("empty expression")
	ErrDivisionByZero  = errors.New("division by zero")
	ErrMissingOperand  = errors.New("missing operand")
	ErrMissingOperator = errors.New("missing operator: too many operands")
	ErrEmptyStack      = errors.New("stack underflow")
)

// SyntaxError — extract position with errors.As
type SyntaxError struct {
	Input    string
	Position int
	Message  string
}

func (e *SyntaxError) Error() string {
	marker := ""
	if e.Position >= 0 && e.Position < len(e.Input) {
		marker = "\n" + e.Input + "\n" + strings.Repeat(" ", e.Position) + "^"
	}

	return fmt.Sprintf("syntax error at position %d: %s%s", e.Position, e.Message, marker)
}

// MismatchError — bracket mismatch with position
type MismatchError struct {
	Input     string
	OpenRune  rune
	OpenPos   int
	CloseRune rune
	ClosePos  int
}

func (e *MismatchError) Error() string {
	if e.ClosePos == -1 {
		return fmt.Sprintf("bracket mismatch: %q at position %d was never closed",
			e.OpenRune, e.OpenPos)
	}
	return fmt.Sprintf("bracket mismatch: %q at position %d closed by %q at position %d",
		e.OpenRune, e.OpenPos, e.CloseRune, e.ClosePos)
}
