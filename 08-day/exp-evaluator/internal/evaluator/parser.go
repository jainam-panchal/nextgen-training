package evaluator

import "exp-evaluator/internal/stack"

func InfixToPostfix(tokens []Token) ([]Token, error) {
	output := make([]Token, 0, len(tokens))
	st := stack.NewStack[Token](len(tokens))

	for _, token := range tokens {
		switch token.Type {
		case TokenNumber:
			output = append(output, token)

		case TokenOperator:
			for !st.IsEmpty() {
				top := st.MustPeek()
				if top.Type != TokenOperator {
					break
				}

				topPrec := OpPrecedence(top.Op)
				curPrec := OpPrecedence(token.Op)
				curAssoc := OpAssoc(token.Op)

				shouldPop := topPrec > curPrec || (topPrec == curPrec && curAssoc == LeftAssoc)
				if !shouldPop {
					break
				}

				output = append(output, st.MustPop())
			}
			st.Push(token)

		case TokenLeftParen:
			st.Push(token)

		case TokenRightParen:
			foundLeft := false
			for !st.IsEmpty() {
				top := st.MustPop()
				if top.Type == TokenLeftParen {
					foundLeft = true
					break
				}
				output = append(output, top)
			}
			if !foundLeft {
				return nil, &MismatchError{
					CloseRune: ')',
					ClosePos:  token.Pos,
					OpenPos:   -1,
				}
			}
		}
	}

	for !st.IsEmpty() {
		top := st.MustPop()
		if top.Type == TokenLeftParen {
			return nil, &MismatchError{
				OpenRune: '(',
				OpenPos:  top.Pos,
				ClosePos: -1,
			}
		}
		output = append(output, top)
	}

	return output, nil
}
