package tree

import "fmt"

var (
	ErrParentNotFound = fmt.Errorf("parent category not found")
	ErrChildExists    = fmt.Errorf("child category already exists")
	ErrNodeNotFound   = fmt.Errorf("category not found")
)

type Node[T comparable] struct {
	Value    T
	Parent   *Node[T]
	Children map[T]*Node[T]
	order    []T
}

type Tree[T comparable] struct {
	Root  *Node[T]
	index map[T]*Node[T]
}

func NewTree[T comparable](root T) *Tree[T] {
	rootNode := &Node[T]{
		Value:    root,
		Children: make(map[T]*Node[T]),
		order:    make([]T, 0),
	}

	return &Tree[T]{
		Root:  rootNode,
		index: map[T]*Node[T]{root: rootNode},
	}
}

func (tree *Tree[T]) Add(parent T, child T) error {
	parentNode, ok := tree.index[parent]
	if !ok {
		return ErrParentNotFound
	}

	if _, exists := tree.index[child]; exists {
		return ErrChildExists
	}

	childNode := &Node[T]{
		Value:    child,
		Parent:   parentNode,
		Children: make(map[T]*Node[T]),
		order:    make([]T, 0),
	}

	parentNode.Children[child] = childNode
	parentNode.order = append(parentNode.order, child)
	tree.index[child] = childNode

	return nil
}

func (tree *Tree[T]) Get(value T) (*Node[T], bool) {
	node, ok := tree.index[value]
	return node, ok
}

func (tree *Tree[T]) SubtreeValues(value T) ([]T, error) {
	startNode, ok := tree.index[value]
	if !ok {
		return nil, ErrNodeNotFound
	}

	queue := []*Node[T]{startNode}
	values := make([]T, 0, len(tree.index))

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		values = append(values, current.Value)

		for _, key := range current.order {
			child := current.Children[key]
			queue = append(queue, child)
		}
	}

	return values, nil
}
