package store

import (
	"testing"
	"time"

	"catalog/internal/models"
)

const benchmarkProductCount = 100_000

func BenchmarkRangeQueryBTreeVsLinear100K(b *testing.B) {
	s := New()
	products := makeProducts(benchmarkProductCount)
	for i := range products {
		if err := s.Create(&products[i]); err != nil {
			b.Fatalf("seed create failed: %v", err)
		}
	}

	cases := []struct {
		name string
		min  float64
		max  float64
	}{
		{name: "narrow", min: 500, max: 700},
		{name: "medium", min: 500, max: 2000},
		{name: "wide", min: 500, max: 20_000},
	}

	for _, tc := range cases {
		tc := tc

		b.Run("BTree/"+tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				min, max := tc.min, tc.max
				_ = s.List(ProductListQuery{
					MinPrice: &min,
					MaxPrice: &max,
					Page:     1,
					Size:     10_000_000,
					Sort:     "price",
					Order:    "asc",
				})
			}
		})

		b.Run("Linear/"+tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = s.linearRangeQuery(tc.min, tc.max)
			}
		})
	}
}

func BenchmarkRangeQueryBTreeProfile1000(b *testing.B) {
	s := New()
	products := makeProducts(benchmarkProductCount)
	for i := range products {
		if err := s.Create(&products[i]); err != nil {
			b.Fatalf("seed create failed: %v", err)
		}
	}

	min, max := 500.0, 2000.0
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < 1000; i++ {
		_ = s.List(ProductListQuery{
			MinPrice: &min,
			MaxPrice: &max,
			Page:     1,
			Size:     10_000_000,
			Sort:     "price",
			Order:    "asc",
		})
	}
}

func BenchmarkRangeQueryLinearProfile1000(b *testing.B) {
	s := New()
	products := makeProducts(benchmarkProductCount)
	for i := range products {
		if err := s.Create(&products[i]); err != nil {
			b.Fatalf("seed create failed: %v", err)
		}
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < 1000; i++ {
		_ = s.linearRangeQuery(500, 2000)
	}
}

func (s *Store) linearRangeQuery(minPrice, maxPrice float64) []*models.Product {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]*models.Product, 0)
	for _, product := range s.byID {
		if product.Price < minPrice || product.Price > maxPrice {
			continue
		}
		productCopy := *product
		out = append(out, &productCopy)
	}
	return out
}

func makeProducts(count int) []models.Product {
	products := make([]models.Product, 0, count)
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	categories := []string{"electronics", "books", "fashion", "home"}

	for i := 0; i < count; i++ {
		price := 100 + float64((i*37)%20000)
		rating := 1 + float64((i%40))/10
		products = append(products, models.Product{
			ID:        "p" + itoa(i),
			Name:      "Product-" + itoa(i),
			Category:  categories[i%len(categories)],
			Price:     price,
			Rating:    rating,
			Stock:     1 + (i % 500),
			Tags:      []string{"tag-a", "tag-b"},
			CreatedAt: baseTime.Add(time.Duration(i) * time.Minute),
		})
	}

	return products
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	digits := [20]byte{}
	i := len(digits)
	for value > 0 {
		i--
		digits[i] = byte('0' + value%10)
		value /= 10
	}
	return string(digits[i:])
}
