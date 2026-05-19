package ranking

type Suggestion struct {
	Word      string `json:"word"`
	Distance  int    `json:"distance"`
	Frequency int    `json:"frequency"`
}

type node struct {
	suggestion Suggestion
	left       *node
	right      *node
}

type BST struct {
	root *node
}

func NewBST() *BST {
	return &BST{}
}

func (tree *BST) Insert(suggestion Suggestion) {
	tree.root = insertNode(tree.root, suggestion)
}

func insertNode(curr *node, suggestion Suggestion) *node {
	if curr == nil {
		return &node{suggestion: suggestion}
	}

	if less(suggestion, curr.suggestion) {
		curr.left = insertNode(curr.left, suggestion)
		return curr
	}

	curr.right = insertNode(curr.right, suggestion)
	return curr
}

func less(first Suggestion, second Suggestion) bool {
	if first.Distance != second.Distance {
		return first.Distance < second.Distance
	}
	if first.Frequency != second.Frequency {
		return first.Frequency > second.Frequency
	}
	return first.Word < second.Word
}

func (tree *BST) Sorted(limit int) []Suggestion {
	if limit <= 0 || tree.root == nil {
		return []Suggestion{}
	}

	results := make([]Suggestion, 0, limit)
	inOrder(tree.root, &results, limit)
	return results
}

func inOrder(curr *node, results *[]Suggestion, limit int) {
	if curr == nil || len(*results) >= limit {
		return
	}

	inOrder(curr.left, results, limit)
	if len(*results) >= limit {
		return
	}

	*results = append(*results, curr.suggestion)
	inOrder(curr.right, results, limit)
}
