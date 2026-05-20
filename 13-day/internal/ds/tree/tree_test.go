package tree

import (
	"errors"
	"slices"
	"testing"
)

func TestTreeAddAndGet(t *testing.T) {
	tree := NewTree("Electronics")

	if err := tree.Add("Electronics", "Phones"); err != nil {
		t.Fatalf("add Phones failed: %v", err)
	}
	if err := tree.Add("Phones", "Smartphones"); err != nil {
		t.Fatalf("add Smartphones failed: %v", err)
	}

	node, ok := tree.Get("Smartphones")
	if !ok {
		t.Fatal("expected Smartphones node to exist")
	}
	if node.Parent == nil || node.Parent.Value != "Phones" {
		t.Fatalf("expected parent Phones, got %+v", node.Parent)
	}
}

func TestTreeAddRejectsInvalidCases(t *testing.T) {
	tree := NewTree("Electronics")

	tests := []struct {
		name      string
		parent    string
		child     string
		setup     func()
		wantError error
	}{
		{
			name:      "missing parent",
			parent:    "NotExists",
			child:     "Phones",
			wantError: ErrParentNotFound,
		},
		{
			name:   "duplicate child",
			parent: "Electronics",
			child:  "Phones",
			setup: func() {
				_ = tree.Add("Electronics", "Phones")
			},
			wantError: ErrChildExists,
		},
		{
			name:   "re-parent existing child",
			parent: "Accessories",
			child:  "Phones",
			setup: func() {
				_ = tree.Add("Electronics", "Phones")
				_ = tree.Add("Electronics", "Accessories")
			},
			wantError: ErrChildExists,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// reset tree per case to keep independence
			tree = NewTree("Electronics")
			if tt.setup != nil {
				tt.setup()
			}

			err := tree.Add(tt.parent, tt.child)
			if !errors.Is(err, tt.wantError) {
				t.Fatalf("expected error %v, got %v", tt.wantError, err)
			}
		})
	}
}

func TestSubtreeValuesBFSOrder(t *testing.T) {
	tree := NewTree("Electronics")
	_ = tree.Add("Electronics", "Phones")
	_ = tree.Add("Electronics", "Laptops")
	_ = tree.Add("Phones", "Smartphones")
	_ = tree.Add("Phones", "FeaturePhones")

	got, err := tree.SubtreeValues("Electronics")
	if err != nil {
		t.Fatalf("SubtreeValues returned error: %v", err)
	}

	// insertion-order BFS
	want := []string{"Electronics", "Phones", "Laptops", "Smartphones", "FeaturePhones"}
	if !slices.Equal(got, want) {
		t.Fatalf("unexpected BFS order: got %v, want %v", got, want)
	}
}

func TestSubtreeValuesMissingNode(t *testing.T) {
	tree := NewTree("Electronics")
	_, err := tree.SubtreeValues("Missing")
	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("expected ErrNodeNotFound, got %v", err)
	}
}
