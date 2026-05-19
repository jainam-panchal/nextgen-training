package btree

import "testing"

func TestBTreeInsertAndSearch(t *testing.T) {
	tree := New[int, string](3)

	tree.Insert(10, "ten")
	tree.Insert(20, "twenty")
	tree.Insert(5, "five")

	value, ok := tree.Search(20)
	if !ok {
		t.Fatal("expected key 20 to exist")
	}

	if value != "twenty" {
		t.Fatalf("expected twenty, got %s", value)
	}
}

func TestBTreeRangeQuery(t *testing.T) {
	tree := New[int, string](3)

	tree.Insert(10, "ten")
	tree.Insert(20, "twenty")
	tree.Insert(5, "five")
	tree.Insert(15, "fifteen")
	tree.Insert(25, "twenty-five")

	values := tree.RangeQuery(10, 20)

	if len(values) != 3 {
		t.Fatalf("expected 3 values, got %d", len(values))
	}
}

func TestBTreeInOrder(t *testing.T) {
	tree := New[int, string](3)

	tree.Insert(30, "thirty")
	tree.Insert(10, "ten")
	tree.Insert(20, "twenty")

	values := tree.InOrder()

	expected := []string{"ten", "twenty", "thirty"}

	if len(values) != len(expected) {
		t.Fatalf("expected %d values, got %d", len(expected), len(values))
	}

	for i := range expected {
		if values[i] != expected[i] {
			t.Fatalf("expected %s at index %d, got %s", expected[i], i, values[i])
		}
	}
}

func TestBTreeDelete(t *testing.T) {
	tree := New[int, string](3)

	tree.Insert(10, "ten")
	tree.Insert(20, "twenty")
	tree.Insert(30, "thirty")

	deleted := tree.Delete(20)
	if !deleted {
		t.Fatal("expected key 20 to be deleted")
	}

	_, ok := tree.Search(20)
	if ok {
		t.Fatal("expected key 20 to be missing after delete")
	}

	if tree.Size() != 2 {
		t.Fatalf("expected size 2, got %d", tree.Size())
	}
}
