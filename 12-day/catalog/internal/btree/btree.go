// Package btree
package btree

import (
	"cmp"
	"fmt"
	"strings"
)

type Node[K cmp.Ordered, V any] struct {
	keys     []K
	values   []V
	children []*Node[K, V]
	leaf     bool
}

type BTree[K cmp.Ordered, V any] struct {
	root *Node[K, V]
	t    int
	size int
}

func New[K cmp.Ordered, V any](minDegree int) *BTree[K, V] {
	if minDegree < 2 {
		panic("min degree must be >= 2")
	}

	return &BTree[K, V]{
		root: &Node[K, V]{leaf: true},
		t:    minDegree,
	}
}

func (tree *BTree[K, V]) Size() int {
	return tree.size
}

func (tree *BTree[K, V]) Search(key K) (V, bool) {
	return search(tree.root, key)
}

func search[K cmp.Ordered, V any](node *Node[K, V], key K) (V, bool) {
	i := 0

	for i < len(node.keys) && key > node.keys[i] {
		i++
	}

	if i < len(node.keys) && key == node.keys[i] {
		return node.values[i], true
	}

	if node.leaf {
		var zero V
		return zero, false
	}

	return search(node.children[i], key)
}

func (tree *BTree[K, V]) Insert(key K, value V) {
	if _, found := tree.Search(key); found {
		tree.update(tree.root, key, value)
		return
	}

	root := tree.root

	if len(root.keys) == 2*tree.t-1 {
		newRoot := &Node[K, V]{
			leaf:     false,
			children: []*Node[K, V]{root},
		}

		tree.splitChild(newRoot, 0)
		tree.root = newRoot
	}

	tree.insertNonFull(tree.root, key, value)
	tree.size++
}

func (tree *BTree[K, V]) update(node *Node[K, V], key K, value V) {
	i := 0

	for i < len(node.keys) && key > node.keys[i] {
		i++
	}

	if i < len(node.keys) && key == node.keys[i] {
		node.values[i] = value
		return
	}

	if !node.leaf {
		tree.update(node.children[i], key, value)
	}
}

func (tree *BTree[K, V]) insertNonFull(node *Node[K, V], key K, value V) {
	i := len(node.keys) - 1

	if node.leaf {
		node.keys = append(node.keys, key)
		node.values = append(node.values, value)

		for i >= 0 && key < node.keys[i] {
			node.keys[i+1] = node.keys[i]
			node.values[i+1] = node.values[i]
			i--
		}

		node.keys[i+1] = key
		node.values[i+1] = value
		return
	}

	for i >= 0 && key < node.keys[i] {
		i--
	}

	i++

	if len(node.children[i].keys) == 2*tree.t-1 {
		tree.splitChild(node, i)

		if key > node.keys[i] {
			i++
		}
	}

	tree.insertNonFull(node.children[i], key, value)
}

func (tree *BTree[K, V]) splitChild(parent *Node[K, V], childIndex int) {
	degree := tree.t
	fullChild := parent.children[childIndex]

	rightChild := &Node[K, V]{
		leaf: fullChild.leaf,
	}

	midKey := fullChild.keys[degree-1]
	midValue := fullChild.values[degree-1]

	rightChild.keys = append(rightChild.keys, fullChild.keys[degree:]...)
	rightChild.values = append(rightChild.values, fullChild.values[degree:]...)

	if !fullChild.leaf {
		rightChild.children = append(rightChild.children, fullChild.children[degree:]...)
		fullChild.children = fullChild.children[:degree]
	}

	fullChild.keys = fullChild.keys[:degree-1]
	fullChild.values = fullChild.values[:degree-1]

	parent.keys = append(parent.keys, midKey)
	parent.values = append(parent.values, midValue)

	copy(parent.keys[childIndex+1:], parent.keys[childIndex:])
	copy(parent.values[childIndex+1:], parent.values[childIndex:])

	parent.keys[childIndex] = midKey
	parent.values[childIndex] = midValue

	parent.children = append(parent.children, rightChild)
	copy(parent.children[childIndex+2:], parent.children[childIndex+1:])
	parent.children[childIndex+1] = rightChild
}

func (tree *BTree[K, V]) RangeQuery(min K, max K) []V {
	result := make([]V, 0)
	rangeQuery(tree.root, min, max, &result)
	return result
}

func rangeQuery[K cmp.Ordered, V any](node *Node[K, V], min K, max K, result *[]V) {
	if node == nil {
		return
	}

	i := 0

	for i < len(node.keys) && node.keys[i] < min {
		if !node.leaf {
			rangeQuery(node.children[i], min, max, result)
		}
		i++
	}

	for i < len(node.keys) {
		if !node.leaf {
			rangeQuery(node.children[i], min, max, result)
		}

		if node.keys[i] > max {
			return
		}

		*result = append(*result, node.values[i])
		i++
	}

	if !node.leaf {
		rangeQuery(node.children[i], min, max, result)
	}
}

func (tree *BTree[K, V]) InOrder() []V {
	result := make([]V, 0, tree.size)
	inOrder(tree.root, &result)
	return result
}

func inOrder[K cmp.Ordered, V any](node *Node[K, V], result *[]V) {
	if node == nil {
		return
	}

	for i := 0; i < len(node.keys); i++ {
		if !node.leaf {
			inOrder(node.children[i], result)
		}

		*result = append(*result, node.values[i])
	}

	if !node.leaf {
		inOrder(node.children[len(node.keys)], result)
	}
}

func (tree *BTree[K, V]) Delete(key K) bool {
	if _, found := tree.Search(key); !found {
		return false
	}

	entries := make([]entry[K, V], 0, tree.size)
	collectEntries(tree.root, &entries)

	tree.root = &Node[K, V]{leaf: true}
	tree.size = 0

	for _, item := range entries {
		if item.key != key {
			tree.Insert(item.key, item.value)
		}
	}

	return true
}

type entry[K cmp.Ordered, V any] struct {
	key   K
	value V
}

func collectEntries[K cmp.Ordered, V any](node *Node[K, V], entries *[]entry[K, V]) {
	if node == nil {
		return
	}

	for i := 0; i < len(node.keys); i++ {
		if !node.leaf {
			collectEntries(node.children[i], entries)
		}

		*entries = append(*entries, entry[K, V]{
			key:   node.keys[i],
			value: node.values[i],
		})
	}

	if !node.leaf {
		collectEntries(node.children[len(node.keys)], entries)
	}
}

func (tree *BTree[K, V]) Visualize() string {
	var builder strings.Builder
	visualize(tree.root, 0, &builder)
	return builder.String()
}

func visualize[K cmp.Ordered, V any](node *Node[K, V], depth int, builder *strings.Builder) {
	if node == nil {
		return
	}

	builder.WriteString(strings.Repeat("  ", depth))
	builder.WriteString(fmt.Sprintf("%v\n", node.keys))

	for _, child := range node.children {
		visualize(child, depth+1, builder)
	}
}
