// Package handlers
package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"catalog/internal/models"
	"catalog/internal/store"
)

type ProductHandler struct {
	store *store.Store
}

type errorResponse struct {
	Error string `json:"error"`
}

func NewProductHandler(s *store.Store) *ProductHandler {
	return &ProductHandler{
		store: s,
	}
}

func (h *ProductHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if id == "" {
		respondError(w, http.StatusBadRequest, "invalid product id provided")
		return
	}

	product, err := h.store.Get(id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondError(w, http.StatusNotFound, "product doesn't exists")
			return
		}

		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, product)
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var p models.Product

	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json "+err.Error())
		return
	}

	if p.ID == "" || p.Name == "" || p.Price < 0 {
		respondError(w, http.StatusBadRequest, "id and name are required; price must be >= 0")
		return
	}

	if err := h.store.Create(&p); err != nil {
		if errors.Is(err, store.ErrDuplicate) {
			respondError(w, http.StatusConflict, "product already exists")
			return
		}

		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, p)
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if id == "" {
		respondError(w, http.StatusBadRequest, "invalid product id provided")
		return
	}

	var p models.Product

	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json "+err.Error())
		return
	}

	if p.Name == "" || p.Price < 0 {
		respondError(w, http.StatusBadRequest, "id and name are required; price must be >= 0")
		return
	}

	if err := h.store.Update(id, &p); err != nil {

		if errors.Is(err, store.ErrNotFound) {
			respondError(w, http.StatusNotFound, "product doesn't exists")
			return
		}

		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if id == "" {
		respondError(w, http.StatusBadRequest, "invalid product id provided")
		return
	}

	if err := h.store.Delete(id); err != nil {

		if errors.Is(err, store.ErrNotFound) {
			respondError(w, http.StatusNotFound, "product doesn't exists")
			return
		}

		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	query, err := parseProductListQuery(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	products := h.store.List(store.ProductListQuery{
		MinPrice: query.MinPrice,
		MaxPrice: query.MaxPrice,
		Category: query.Category,
		Sort:     query.Sort,
		Order:    query.Order,
		Page:     query.Page,
		Size:     query.Size,
	})

	respondJSON(w, http.StatusOK, products)
}

func parseProductListQuery(r *http.Request) (store.ProductListQuery, error) {
	values := r.URL.Query()

	query := store.ProductListQuery{
		Category: values.Get("category"),
		Sort:     values.Get("sort"),
		Order:    values.Get("order"),
		Page:     1,
		Size:     20,
	}

	if query.Sort == "" {
		query.Sort = "price"
	}

	if query.Order == "" {
		query.Order = "asc"
	}

	if query.Order != "asc" && query.Order != "desc" {
		return store.ProductListQuery{}, errors.New("order must be either asc or desc")
	}

	switch query.Sort {
	case "price", "rating", "name", "created_at":
	default:
		return store.ProductListQuery{}, errors.New("sort must be one of: price, rating, name, created_at")
	}

	if minPriceRaw := values.Get("min_price"); minPriceRaw != "" {
		minPrice, err := strconv.ParseFloat(minPriceRaw, 64)
		if err != nil {
			return store.ProductListQuery{}, errors.New("min_price must be a valid number")
		}

		if minPrice < 0 {
			return store.ProductListQuery{}, errors.New("min_price must be >= 0")
		}

		query.MinPrice = &minPrice
	}

	if maxPriceRaw := values.Get("max_price"); maxPriceRaw != "" {
		maxPrice, err := strconv.ParseFloat(maxPriceRaw, 64)
		if err != nil {
			return store.ProductListQuery{}, errors.New("max_price must be a valid number")
		}

		if maxPrice < 0 {
			return store.ProductListQuery{}, errors.New("max_price must be >= 0")
		}

		query.MaxPrice = &maxPrice
	}

	if query.MinPrice != nil && query.MaxPrice != nil && *query.MinPrice > *query.MaxPrice {
		return store.ProductListQuery{}, errors.New("min_price cannot be greater than max_price")
	}

	if pageRaw := values.Get("page"); pageRaw != "" {
		page, err := strconv.Atoi(pageRaw)
		if err != nil {
			return store.ProductListQuery{}, errors.New("page must be a valid integer")
		}

		if page <= 0 {
			return store.ProductListQuery{}, errors.New("page must be greater than 0")
		}

		query.Page = page
	}

	if sizeRaw := values.Get("size"); sizeRaw != "" {
		size, err := strconv.Atoi(sizeRaw)
		if err != nil {
			return store.ProductListQuery{}, errors.New("size must be a valid integer")
		}

		if size <= 0 {
			return store.ProductListQuery{}, errors.New("size must be greater than 0")
		}

		if size > 100 {
			return store.ProductListQuery{}, errors.New("size cannot be greater than 100")
		}

		query.Size = size
	}

	return query, nil
}

func respondJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, errorResponse{Error: message})
}
