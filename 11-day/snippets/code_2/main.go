package trie

type Trie struct {
	root      *TrieNode
	wordCount int
}

func (t *Trie) NewTrie() *Trie {
	return &Trie{
		root:      NewTrieNode(),
		wordCount: 0,
	}
}

type TrieNode struct {
	children  map[rune]*TrieNode
	isWord    bool
	frequency int
}

func NewTrieNode() *TrieNode {
	return &TrieNode{
		children: make(map[rune]*TrieNode),
	}
}

func (t *Trie) Insert(word string, freq int) {

	if word == "" {
		return
	}

	node := t.root
	for _, ch := range word {
		if _, ok := node.children[ch]; !ok {
			node.children[ch] = NewTrieNode()
		}
		node = node.children[ch]
	}

	if !node.isWord {
		t.wordCount++
	}
	node.isWord = true
	node.frequency += freq

}

func (t *Trie) Search(word string) bool {
	if word == "" {
		return false
	}

	node := t.root
	for _, ch := range word {
		if _, ok := node.children[ch]; !ok {
			return false
		}
		node = node.children[ch]
	}

	return node.isWord
}
