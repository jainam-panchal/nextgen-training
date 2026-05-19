package trie

type Trie struct {
	root *Node
}

type Node struct {
	Children  map[rune]*Node
	IsWord    bool
	Frequency int
}

type Suggestion struct {
	Word      string
	Frequency int
}

func NewTrie() *Trie {
	return &Trie{
		root: &Node{
			Children: make(map[rune]*Node),
		},
	}
}

func (t *Trie) Insert(word string, frequency int) {
	if word == "" {
		return
	}

	curr := t.root

	for _, r := range word {
		if _, ok := curr.Children[r]; !ok {
			curr.Children[r] = &Node{
				Children: make(map[rune]*Node),
			}
		}

		curr = curr.Children[r]
	}

	curr.IsWord = true
	curr.Frequency += frequency
}

func (t *Trie) Search(word string) bool {
	if word == "" {
		return false
	}
	curr := t.root
	for _, r := range word {
		if _, ok := curr.Children[r]; !ok {
			return false
		}
		curr = curr.Children[r]
	}
	return curr.IsWord
}

func (t *Trie) StartsWith(prefix string) bool {
	if prefix == "" {
		return false
	}

	curr := t.root
	for _, r := range prefix {
		if _, ok := curr.Children[r]; !ok {
			return false
		}
		curr = curr.Children[r]
	}

	return true
}

func (t *Trie) Frequency(word string) int {
	if word == "" {
		return 0
	}
	curr := t.root
	for _, r := range word {
		next, ok := curr.Children[r]
		if !ok {
			return 0
		}

		curr = next
	}
	if !curr.IsWord {
		return 0
	}
	return curr.Frequency
}

// helper
func (t *Trie) findNode(prefix string) *Node {
	if prefix == "" {
		return nil
	}

	curr := t.root
	for _, r := range prefix {
		next, ok := curr.Children[r]
		if !ok {
			return nil
		}
		curr = next
	}

	return curr
}

func (t *Trie) AutoComplete(prefix string, limit int) []Suggestion {
	if prefix == "" || limit <= 0 {
		return []Suggestion{}
	}

	node := t.findNode(prefix)
	if node == nil {
		return []Suggestion{}
	}

	results := make([]Suggestion, 0, limit)
	collectTopWords(node, []rune(prefix), &results, limit)
	return results
}

func (t *Trie) Delete(word string) bool {
	if word == "" {
		return false
	}

	_, deleted := deleteWord(t.root, []rune(word), 0)
	return deleted
}

func deleteWord(node *Node, runes []rune, index int) (shouldPrune bool, deleted bool) {
	if index == len(runes) {
		if !node.IsWord {
			return false, false
		}

		node.IsWord = false
		node.Frequency = 0

		return len(node.Children) == 0, true
	}

	r := runes[index]
	child, ok := node.Children[r]
	if !ok {
		return false, false
	}

	childShouldPrune, deleted := deleteWord(child, runes, index+1)
	if childShouldPrune {
		delete(node.Children, r)
	}

	shouldPrune = !node.IsWord && len(node.Children) == 0
	return shouldPrune, deleted
}

func collectWords(node *Node, currentWord []rune, results *[]Suggestion) {
	if node == nil {
		return
	}
	if node.IsWord {
		*results = append(*results, Suggestion{
			Word:      string(currentWord),
			Frequency: node.Frequency,
		})
	}
	for r, child := range node.Children {
		currentWord = append(currentWord, r)
		collectWords(child, currentWord, results)
		currentWord = currentWord[:len(currentWord)-1]
	}
}

func collectTopWords(node *Node, currentWord []rune, results *[]Suggestion, limit int) {
	if node == nil {
		return
	}

	if node.IsWord {
		insertTopSuggestion(results, Suggestion{
			Word:      string(currentWord),
			Frequency: node.Frequency,
		}, limit)
	}

	for r, child := range node.Children {
		currentWord = append(currentWord, r)
		collectTopWords(child, currentWord, results, limit)
		currentWord = currentWord[:len(currentWord)-1]
	}
}

func insertTopSuggestion(results *[]Suggestion, candidate Suggestion, limit int) {
	list := *results
	insertAt := len(list)
	for i := range list {
		if betterSuggestion(candidate, list[i]) {
			insertAt = i
			break
		}
	}

	if insertAt == len(list) {
		if len(list) < limit {
			*results = append(list, candidate)
		}
		return
	}

	list = append(list, Suggestion{})
	copy(list[insertAt+1:], list[insertAt:])
	list[insertAt] = candidate

	if len(list) > limit {
		list = list[:limit]
	}

	*results = list
}

func betterSuggestion(left, right Suggestion) bool {
	if left.Frequency == right.Frequency {
		return left.Word < right.Word
	}
	return left.Frequency > right.Frequency
}
