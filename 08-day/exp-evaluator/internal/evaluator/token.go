package evaluator

import "fmt"

type TokenType int

const (
	TokenNumber TokenType = iota
	TokenOperator
	TokenLeftParen
	TokenRightParen
)

type Assoc int

const (
	LeftAssoc Assoc = iota
	RightAssoc
)

type OpMeta struct {
	Precedence int
	Assoc      Assoc
}

var ops = map[rune]OpMeta{
	'+': {1, LeftAssoc},
	'-': {1, LeftAssoc},
	'*': {2, LeftAssoc},
	'/': {2, LeftAssoc},
	'^': {3, RightAssoc},
}

func IsOperator(r rune) bool {
	_, ok := ops[r]
	return ok
}

func OpPrecedence(r rune) int {
	return ops[r].Precedence
}

func OpAssoc(r rune) Assoc {
	return ops[r].Assoc
}

type Token struct {
	Type  TokenType
	Value float64 
	Op    rune    // valid when Type == TokenOperator
	Pos   int     // position in original input (for error messages)
}

func (t Token) String() string {
	switch t.Type {
	case TokenNumber:
		return fmt.Sprintf("%.6g", t.Value)
	case TokenOperator:
		return string(t.Op)
	case TokenLeftParen:
		return "("
	case TokenRightParen:
		return ")"
	}
	return "?"
}
