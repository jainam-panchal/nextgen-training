package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"catalog/internal/models"
	"catalog/internal/store"
)

func TestUpdateInvalidJSONReturns400(t *testing.T) {
	s := store.New()
	_ = s.Create(&models.Product{
		ID:       "p1",
		Name:     "Phone",
		Category: "electronics",
		Price:    100,
		Rating:   4.5,
	})

	h := NewProductHandler(s)

	req := httptest.NewRequest(http.MethodPut, "/products/p1", strings.NewReader("{invalid"))
	req.SetPathValue("id", "p1")
	rec := httptest.NewRecorder()

	h.Update(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestUpdateSuccessReturns204NoContent(t *testing.T) {
	s := store.New()
	_ = s.Create(&models.Product{
		ID:       "p1",
		Name:     "Phone",
		Category: "electronics",
		Price:    100,
		Rating:   4.5,
	})

	h := NewProductHandler(s)

	reqBody := `{"name":"Phone Pro","category":"electronics","price":120,"rating":4.7}`
	req := httptest.NewRequest(http.MethodPut, "/products/p1", strings.NewReader(reqBody))
	req.SetPathValue("id", "p1")
	rec := httptest.NewRecorder()

	h.Update(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("expected empty body for 204, got %q", rec.Body.String())
	}
}

func TestDeleteSuccessReturns204NoContent(t *testing.T) {
	s := store.New()
	_ = s.Create(&models.Product{
		ID:       "p1",
		Name:     "Phone",
		Category: "electronics",
		Price:    100,
		Rating:   4.5,
	})

	h := NewProductHandler(s)

	req := httptest.NewRequest(http.MethodDelete, "/products/p1", nil)
	req.SetPathValue("id", "p1")
	rec := httptest.NewRecorder()

	h.Delete(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("expected empty body for 204, got %q", rec.Body.String())
	}
}
