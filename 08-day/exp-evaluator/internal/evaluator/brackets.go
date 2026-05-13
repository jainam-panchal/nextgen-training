package evaluator

import (
	"exp-evaluator/internal/stack"
)

func BracketMatch(expr string) error {
	type openBracket struct {
		r   rune
		pos int
	}

	pairs := map[rune]rune{
		')': '(',
		']': '[',
		'}': '{',
		'>': '<',
	}

	isOpen := map[rune]bool{
		'(': true,
		'[': true,
		'{': true,
	}

	st := stack.NewStack[openBracket](len(expr))

	for pos, ch := range expr {
		if isOpen[ch] {
			st.Push(openBracket{r: ch, pos: pos})
			continue
		}

		expected, isClose := pairs[ch]
		if !isClose {
			continue // not a bracket
		}

		// closing bracket without opening bracket
		if st.IsEmpty() {
			return &MismatchError{CloseRune: ch, ClosePos: pos, OpenPos: -1}
		}

		top, _ := st.Pop()
		if top.r != expected {
			return &MismatchError{
				OpenRune:  top.r,
				OpenPos:   top.pos,
				ClosePos:  pos,
				CloseRune: ch,
			}
		}
	}

	if !st.IsEmpty() {
		unclosed, _ := st.Pop()
		return &MismatchError{
			OpenRune: unclosed.r,
			OpenPos:  unclosed.pos,
			ClosePos: -1,
		}
	}

	return nil

}
