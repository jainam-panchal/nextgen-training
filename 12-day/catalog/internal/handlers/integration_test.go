package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"catalog/internal/middleware"
	"catalog/internal/models"
	"catalog/internal/store"
)

func TestAPIIntegration_EndToEnd(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	createdAt := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	createProduct(t, server, models.Product{
		ID:        "p1",
		Name:      "Keyboard",
		Category:  "electronics",
		Price:     1500,
		Rating:    4.6,
		Stock:     20,
		Tags:      []string{"input"},
		CreatedAt: createdAt,
	})
	createProduct(t, server, models.Product{
		ID:        "p2",
		Name:      "Book A",
		Category:  "books",
		Price:     700,
		Rating:    4.8,
		Stock:     50,
		Tags:      []string{"learning"},
		CreatedAt: createdAt.Add(time.Hour),
	})

	t.Run("GET by ID", func(t *testing.T) {
		res := mustRequest(t, http.MethodGet, server.URL+"/products/p1", nil)
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", res.StatusCode)
		}
	})

	t.Run("Range query by price", func(t *testing.T) {
		res := mustRequest(t, http.MethodGet, server.URL+"/products?min_price=500&max_price=1000", nil)
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", res.StatusCode)
		}
		var products []models.Product
		_ = json.NewDecoder(res.Body).Decode(&products)
		if len(products) != 1 || products[0].ID != "p2" {
			t.Fatalf("expected only p2 in range, got %+v", products)
		}
	})

	t.Run("Category + sort", func(t *testing.T) {
		createProduct(t, server, models.Product{
			ID:        "p3",
			Name:      "Mouse",
			Category:  "electronics",
			Price:     500,
			Rating:    4.1,
			Stock:     30,
			CreatedAt: createdAt.Add(2 * time.Hour),
		})
		res := mustRequest(t, http.MethodGet, server.URL+"/products?category=electronics&sort=price&order=asc", nil)
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", res.StatusCode)
		}
		var products []models.Product
		_ = json.NewDecoder(res.Body).Decode(&products)
		if len(products) < 2 || products[0].ID != "p3" {
			t.Fatalf("expected p3 first after ascending sort, got %+v", products)
		}
	})

	t.Run("Pagination", func(t *testing.T) {
		res := mustRequest(t, http.MethodGet, server.URL+"/products?page=1&size=2&sort=price&order=asc", nil)
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", res.StatusCode)
		}
		var products []models.Product
		_ = json.NewDecoder(res.Body).Decode(&products)
		if len(products) != 2 {
			t.Fatalf("expected 2 products, got %d", len(products))
		}
	})

	t.Run("Update", func(t *testing.T) {
		update := models.Product{
			Name:      "Keyboard Pro",
			Category:  "electronics",
			Price:     1700,
			Rating:    4.7,
			Stock:     22,
			Tags:      []string{"input", "pro"},
			CreatedAt: createdAt,
		}
		body, _ := json.Marshal(update)
		res := mustRequest(t, http.MethodPut, server.URL+"/products/p1", bytes.NewReader(body))
		defer res.Body.Close()
		if res.StatusCode != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", res.StatusCode)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		res := mustRequest(t, http.MethodDelete, server.URL+"/products/p2", nil)
		defer res.Body.Close()
		if res.StatusCode != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", res.StatusCode)
		}
	})

	t.Run("Stats", func(t *testing.T) {
		res := mustRequest(t, http.MethodGet, server.URL+"/products/stats", nil)
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", res.StatusCode)
		}
		var stats map[string]any
		_ = json.NewDecoder(res.Body).Decode(&stats)
		if _, ok := stats["total_products"]; !ok {
			t.Fatalf("expected total_products in stats, got %+v", stats)
		}
	})
}

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	s := store.New()
	h := NewProductHandler(s)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /products", h.Create)
	mux.HandleFunc("GET /products", h.List)
	mux.HandleFunc("GET /products/stats", h.Stats)
	mux.HandleFunc("GET /products/{id}", h.Get)
	mux.HandleFunc("PUT /products/{id}", h.Update)
	mux.HandleFunc("DELETE /products/{id}", h.Delete)

	handler := middleware.Logging(
		middleware.Recovery(
			middleware.JSONOnly(
				middleware.Timing(mux),
			),
		),
	)

	return httptest.NewServer(handler)
}

func createProduct(t *testing.T, server *httptest.Server, product models.Product) {
	t.Helper()
	body, _ := json.Marshal(product)
	res := mustRequest(t, http.MethodPost, server.URL+"/products", bytes.NewReader(body))
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 create, got %d", res.StatusCode)
	}
}

func mustRequest(t *testing.T, method, url string, body *bytes.Reader) *http.Response {
	t.Helper()

	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		reader = body
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatalf("new request failed: %v", err)
	}
	if method == http.MethodPost || method == http.MethodPut {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	return res
}
