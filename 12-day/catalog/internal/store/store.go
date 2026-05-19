// Package store
package store

import (
	"errors"
	"sort"
	"sync"

	"catalog/internal/btree"
	"catalog/internal/models"
)

var (
	ErrNotFound  = errors.New("product not found")
	ErrDuplicate = errors.New("product already exists")
)

type ProductListQuery struct {
	MinPrice *float64
	MaxPrice *float64
	Category string
	Sort     string
	Order    string
	Page     int
	Size     int
}

type Stats struct {
	TotalProducts  int            `json:"total_products"`
	CategoryCounts map[string]int `json:"category_counts"`
	PriceMin       float64        `json:"price_min"`
	PriceMax       float64        `json:"price_max"`
	AvgRating      float64        `json:"avg_rating"`
}

type Store struct {
	mu      sync.RWMutex
	byID    map[string]*models.Product
	byPrice *btree.BTree[float64, []*models.Product]
}

func New() *Store {
	return &Store{
		byID:    make(map[string]*models.Product),
		byPrice: btree.New[float64, []*models.Product](32),
	}
}

func (s *Store) Create(product *models.Product) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if product == nil {
		return errors.New("product cannot be nil")
	}

	if _, exists := s.byID[product.ID]; exists {
		return ErrDuplicate
	}

	productCopy := *product

	s.byID[productCopy.ID] = &productCopy
	s.addToPriceIndex(&productCopy)

	return nil
}

func (s *Store) Get(id string) (*models.Product, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	product, exists := s.byID[id]
	if !exists {
		return nil, ErrNotFound
	}

	productCopy := *product
	return &productCopy, nil
}

func (s *Store) Update(id string, product *models.Product) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if product == nil {
		return errors.New("product cannot be nil")
	}

	existingProduct, exists := s.byID[id]
	if !exists {
		return ErrNotFound
	}

	s.removeFromPriceIndex(existingProduct)

	productCopy := *product
	productCopy.ID = id

	s.byID[id] = &productCopy
	s.addToPriceIndex(&productCopy)

	return nil
}

func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	product, exists := s.byID[id]
	if !exists {
		return ErrNotFound
	}

	s.removeFromPriceIndex(product)
	delete(s.byID, id)

	return nil
}

func (s *Store) List(query ProductListQuery) []*models.Product {
	s.mu.RLock()
	defer s.mu.RUnlock()

	candidates := s.getListCandidates(query)

	filteredProducts := filterProducts(candidates, query)

	sortProducts(filteredProducts, query.Sort, query.Order)

	return paginateProducts(filteredProducts, query.Page, query.Size)
}

func (s *Store) Stats() Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := Stats{
		TotalProducts:  len(s.byID),
		CategoryCounts: make(map[string]int),
	}

	if len(s.byID) == 0 {
		return stats
	}

	var totalRating float64
	firstProduct := true

	for _, product := range s.byID {
		stats.CategoryCounts[product.Category]++
		totalRating += product.Rating

		if firstProduct {
			stats.PriceMin = product.Price
			stats.PriceMax = product.Price
			firstProduct = false
			continue
		}

		if product.Price < stats.PriceMin {
			stats.PriceMin = product.Price
		}

		if product.Price > stats.PriceMax {
			stats.PriceMax = product.Price
		}
	}

	stats.AvgRating = totalRating / float64(len(s.byID))

	return stats
}

func (s *Store) getListCandidates(query ProductListQuery) []*models.Product {
	hasPriceFilter := query.MinPrice != nil || query.MaxPrice != nil

	if hasPriceFilter {
		return s.getCandidatesFromPriceIndex(query)
	}

	return s.getCandidatesFromMap()
}

func (s *Store) getCandidatesFromPriceIndex(query ProductListQuery) []*models.Product {
	minPrice := 0.0
	maxPrice := 1e18

	if query.MinPrice != nil {
		minPrice = *query.MinPrice
	}

	if query.MaxPrice != nil {
		maxPrice = *query.MaxPrice
	}

	groups := s.byPrice.RangeQuery(minPrice, maxPrice)

	var products []*models.Product

	for _, group := range groups {
		products = append(products, group...)
	}

	return products
}

func (s *Store) getCandidatesFromMap() []*models.Product {
	products := make([]*models.Product, 0, len(s.byID))

	for _, product := range s.byID {
		products = append(products, product)
	}

	return products
}

func filterProducts(products []*models.Product, query ProductListQuery) []*models.Product {
	filteredProducts := make([]*models.Product, 0, len(products))

	for _, product := range products {
		if query.Category != "" && product.Category != query.Category {
			continue
		}

		productCopy := *product
		filteredProducts = append(filteredProducts, &productCopy)
	}

	return filteredProducts
}

func sortProducts(products []*models.Product, sortBy string, order string) {
	if sortBy == "" {
		sortBy = "price"
	}

	descending := order == "desc"

	sort.Slice(products, func(i int, j int) bool {
		var less bool

		switch sortBy {
		case "price":
			less = products[i].Price < products[j].Price
		case "rating":
			less = products[i].Rating < products[j].Rating
		case "name":
			less = products[i].Name < products[j].Name
		case "created_at":
			less = products[i].CreatedAt.Before(products[j].CreatedAt)
		default:
			less = products[i].Price < products[j].Price
		}

		if descending {
			switch sortBy {
			case "price":
				return products[i].Price > products[j].Price
			case "rating":
				return products[i].Rating > products[j].Rating
			case "name":
				return products[i].Name > products[j].Name
			case "created_at":
				return products[i].CreatedAt.After(products[j].CreatedAt)
			default:
				return products[i].Price > products[j].Price
			}
		}

		return less
	})
}

func paginateProducts(products []*models.Product, page int, size int) []*models.Product {
	if page <= 0 {
		page = 1
	}

	if size <= 0 {
		size = 20
	}

	offset := (page - 1) * size

	if offset >= len(products) {
		return []*models.Product{}
	}

	end := offset + size
	if end > len(products) {
		end = len(products)
	}

	return products[offset:end]
}

func (s *Store) addToPriceIndex(product *models.Product) {
	group, exists := s.byPrice.Search(product.Price)
	if !exists {
		s.byPrice.Insert(product.Price, []*models.Product{product})
		return
	}

	group = append(group, product)
	s.byPrice.Insert(product.Price, group)
}

func (s *Store) removeFromPriceIndex(product *models.Product) {
	group, exists := s.byPrice.Search(product.Price)
	if !exists {
		return
	}

	nextGroup := make([]*models.Product, 0, len(group))

	for _, item := range group {
		if item.ID != product.ID {
			nextGroup = append(nextGroup, item)
		}
	}

	if len(nextGroup) == 0 {
		s.byPrice.Delete(product.Price)
		return
	}

	s.byPrice.Insert(product.Price, nextGroup)
}
