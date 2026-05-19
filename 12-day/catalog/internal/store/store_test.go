package store

import (
	"math"
	"testing"
	"time"

	"catalog/internal/models"
)

func TestCreateAndGetProduct(t *testing.T) {
	s := New()

	product := &models.Product{
		ID:       "p1",
		Name:     "Keyboard",
		Category: "electronics",
		Price:    1500,
		Rating:   4.5,
		Stock:    10,
	}

	if err := s.Create(product); err != nil {
		t.Fatalf("expected create to succeed, got %v", err)
	}

	got, err := s.Get("p1")
	if err != nil {
		t.Fatalf("expected get to succeed, got %v", err)
	}

	if got.Name != "Keyboard" {
		t.Fatalf("expected Keyboard, got %s", got.Name)
	}
}

func TestGetMissingProduct(t *testing.T) {
	s := New()

	_, err := s.Get("missing")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestCreateDuplicateProduct(t *testing.T) {
	s := New()

	product := &models.Product{
		ID:    "p1",
		Name:  "Mouse",
		Price: 500,
	}

	if err := s.Create(product); err != nil {
		t.Fatalf("expected first create to succeed, got %v", err)
	}

	err := s.Create(product)
	if err != ErrDuplicate {
		t.Fatalf("expected ErrDuplicate, got %v", err)
	}
}

func TestListByPriceRange(t *testing.T) {
	s := New()

	seedProducts(t, s)

	minPrice := 400.0
	maxPrice := 2000.0

	got := s.List(ProductListQuery{
		MinPrice: &minPrice,
		MaxPrice: &maxPrice,
		Page:     1,
		Size:     20,
		Sort:     "price",
		Order:    "asc",
	})

	if len(got) != 3 {
		t.Fatalf("expected 3 products, got %d", len(got))
	}

	expectedIDs := []string{"p1", "p2", "p4"}

	for i, expectedID := range expectedIDs {
		if got[i].ID != expectedID {
			t.Fatalf("expected product %s at index %d, got %s", expectedID, i, got[i].ID)
		}
	}
}

func TestListByCategory(t *testing.T) {
	s := New()

	seedProducts(t, s)

	got := s.List(ProductListQuery{
		Category: "electronics",
		Page:     1,
		Size:     20,
		Sort:     "price",
		Order:    "asc",
	})

	if len(got) != 3 {
		t.Fatalf("expected 3 electronics products, got %d", len(got))
	}

	for _, product := range got {
		if product.Category != "electronics" {
			t.Fatalf("expected electronics category, got %s", product.Category)
		}
	}
}

func TestListByCategoryAndPriceRange(t *testing.T) {
	s := New()

	seedProducts(t, s)

	minPrice := 1000.0
	maxPrice := 5000.0

	got := s.List(ProductListQuery{
		MinPrice: &minPrice,
		MaxPrice: &maxPrice,
		Category: "electronics",
		Page:     1,
		Size:     20,
		Sort:     "price",
		Order:    "asc",
	})

	if len(got) != 1 {
		t.Fatalf("expected 1 product, got %d", len(got))
	}

	if got[0].ID != "p2" {
		t.Fatalf("expected p2, got %s", got[0].ID)
	}
}

func TestListPagination(t *testing.T) {
	s := New()

	seedProducts(t, s)

	got := s.List(ProductListQuery{
		Page:  2,
		Size:  2,
		Sort:  "price",
		Order: "asc",
	})

	if len(got) != 2 {
		t.Fatalf("expected 2 products, got %d", len(got))
	}

	if got[0].ID != "p4" {
		t.Fatalf("expected first product on page 2 to be p4, got %s", got[0].ID)
	}

	if got[1].ID != "p3" {
		t.Fatalf("expected second product on page 2 to be p3, got %s", got[1].ID)
	}
}

func TestListSortByPriceDescending(t *testing.T) {
	s := New()

	seedProducts(t, s)

	got := s.List(ProductListQuery{
		Page:  1,
		Size:  20,
		Sort:  "price",
		Order: "desc",
	})

	if len(got) == 0 {
		t.Fatal("expected products")
	}

	if got[0].ID != "p3" {
		t.Fatalf("expected highest priced product p3, got %s", got[0].ID)
	}
}

func TestListSortByRatingDescending(t *testing.T) {
	s := New()

	seedProducts(t, s)

	got := s.List(ProductListQuery{
		Page:  1,
		Size:  20,
		Sort:  "rating",
		Order: "desc",
	})

	if len(got) == 0 {
		t.Fatal("expected products")
	}

	if got[0].ID != "p4" {
		t.Fatalf("expected highest rated product p4, got %s", got[0].ID)
	}
}

func TestUpdateProduct(t *testing.T) {
	s := New()

	product := &models.Product{
		ID:       "p1",
		Name:     "Mouse",
		Category: "electronics",
		Price:    500,
	}

	if err := s.Create(product); err != nil {
		t.Fatalf("create failed: %v", err)
	}

	updated := &models.Product{
		ID:       "p1",
		Name:     "Gaming Mouse",
		Category: "electronics",
		Price:    900,
	}

	if err := s.Update("p1", updated); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	got, err := s.Get("p1")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}

	if got.Name != "Gaming Mouse" {
		t.Fatalf("expected Gaming Mouse, got %s", got.Name)
	}

	oldPrice := 500.0
	oldPriceResults := s.List(ProductListQuery{
		MinPrice: &oldPrice,
		MaxPrice: &oldPrice,
		Page:     1,
		Size:     20,
		Sort:     "price",
		Order:    "asc",
	})

	if len(oldPriceResults) != 0 {
		t.Fatalf("expected old price index to be empty, got %d", len(oldPriceResults))
	}

	newPrice := 900.0
	newPriceResults := s.List(ProductListQuery{
		MinPrice: &newPrice,
		MaxPrice: &newPrice,
		Page:     1,
		Size:     20,
		Sort:     "price",
		Order:    "asc",
	})

	if len(newPriceResults) != 1 {
		t.Fatalf("expected new price index to contain product, got %d", len(newPriceResults))
	}
}

func TestUpdateMissingProduct(t *testing.T) {
	s := New()

	product := &models.Product{
		ID:    "missing",
		Name:  "Missing Product",
		Price: 100,
	}

	err := s.Update("missing", product)
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDeleteProduct(t *testing.T) {
	s := New()

	product := &models.Product{
		ID:       "p1",
		Name:     "Mouse",
		Category: "electronics",
		Price:    500,
	}

	if err := s.Create(product); err != nil {
		t.Fatalf("create failed: %v", err)
	}

	if err := s.Delete("p1"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	_, err := s.Get("p1")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	price := 500.0
	results := s.List(ProductListQuery{
		MinPrice: &price,
		MaxPrice: &price,
		Page:     1,
		Size:     20,
		Sort:     "price",
		Order:    "asc",
	})

	if len(results) != 0 {
		t.Fatalf("expected deleted product to be removed from price index, got %d results", len(results))
	}
}

func TestDeleteMissingProduct(t *testing.T) {
	s := New()

	err := s.Delete("missing")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestStats(t *testing.T) {
	s := New()

	seedProducts(t, s)

	stats := s.Stats()

	if stats.TotalProducts != 4 {
		t.Fatalf("expected 4 products, got %d", stats.TotalProducts)
	}

	if stats.CategoryCounts["electronics"] != 3 {
		t.Fatalf("expected 3 electronics products, got %d", stats.CategoryCounts["electronics"])
	}

	if stats.CategoryCounts["books"] != 1 {
		t.Fatalf("expected 1 books product, got %d", stats.CategoryCounts["books"])
	}

	if stats.PriceMin != 500 {
		t.Fatalf("expected min price 500, got %f", stats.PriceMin)
	}

	if stats.PriceMax != 8000 {
		t.Fatalf("expected max price 8000, got %f", stats.PriceMax)
	}

	expectedAvgRating := (4.2 + 4.5 + 4.1 + 4.9) / 4
	if math.Abs(stats.AvgRating-expectedAvgRating) > 1e-9 {
		t.Fatalf("expected avg rating %f, got %f", expectedAvgRating, stats.AvgRating)
	}
}

func TestGetReturnsCopy(t *testing.T) {
	s := New()

	product := &models.Product{
		ID:       "p1",
		Name:     "Mouse",
		Category: "electronics",
		Price:    500,
	}

	if err := s.Create(product); err != nil {
		t.Fatalf("create failed: %v", err)
	}

	got, err := s.Get("p1")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}

	got.Name = "Mutated Mouse"
	got.Price = 999999

	again, err := s.Get("p1")
	if err != nil {
		t.Fatalf("second get failed: %v", err)
	}

	if again.Name != "Mouse" {
		t.Fatalf("expected stored product name to remain Mouse, got %s", again.Name)
	}

	if again.Price != 500 {
		t.Fatalf("expected stored product price to remain 500, got %f", again.Price)
	}
}

func TestNilProductSafety(t *testing.T) {
	s := New()

	if err := s.Create(nil); err == nil {
		t.Fatal("expected create(nil) to return error")
	}

	if err := s.Update("p1", nil); err == nil {
		t.Fatal("expected update(nil) to return error")
	}
}

func seedProducts(t *testing.T, s *Store) {
	t.Helper()

	products := []*models.Product{
		{
			ID:        "p1",
			Name:      "Mouse",
			Category:  "electronics",
			Price:     500,
			Rating:    4.2,
			Stock:     30,
			CreatedAt: time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC),
		},
		{
			ID:        "p2",
			Name:      "Keyboard",
			Category:  "electronics",
			Price:     1500,
			Rating:    4.5,
			Stock:     20,
			CreatedAt: time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC),
		},
		{
			ID:        "p3",
			Name:      "Monitor",
			Category:  "electronics",
			Price:     8000,
			Rating:    4.1,
			Stock:     15,
			CreatedAt: time.Date(2026, 1, 3, 10, 0, 0, 0, time.UTC),
		},
		{
			ID:        "p4",
			Name:      "Book",
			Category:  "books",
			Price:     2000,
			Rating:    4.9,
			Stock:     50,
			CreatedAt: time.Date(2026, 1, 4, 10, 0, 0, 0, time.UTC),
		},
	}

	for _, product := range products {
		if err := s.Create(product); err != nil {
			t.Fatalf("seed create failed for %s: %v", product.ID, err)
		}
	}
}
