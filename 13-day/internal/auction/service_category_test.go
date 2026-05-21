package auction

import (
	"errors"
	"testing"
	"time"

	"realtime-auction/internal/models"
)

func TestBrowseCategoryIncludesDescendants(t *testing.T) {
	store := NewAuctionStore()
	service := NewAuctionService(store)

	sellerID := mustSeedUser(t, store, "seller", 1000)
	_, err := service.CreateItem(
		"Phone 1",
		[]string{"Electronics", "Phones", "Smartphones"},
		"desc",
		sellerID,
		100,
		time.Now().Add(-time.Hour),
		time.Now().Add(time.Hour),
	)
	if err != nil {
		t.Fatalf("CreateItem failed: %v", err)
	}
	_, err = service.CreateItem(
		"Laptop 1",
		[]string{"Electronics", "Laptops"},
		"desc",
		sellerID,
		100,
		time.Now().Add(-time.Hour),
		time.Now().Add(time.Hour),
	)
	if err != nil {
		t.Fatalf("CreateItem failed: %v", err)
	}

	items, err := service.BrowseCategory("Electronics")
	if err != nil {
		t.Fatalf("BrowseCategory failed: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items in Electronics subtree, got %d", len(items))
	}
}

func TestBrowseCategoryExactLeaf(t *testing.T) {
	store := NewAuctionStore()
	service := NewAuctionService(store)

	sellerID := mustSeedUser(t, store, "seller", 1000)
	_, err := service.CreateItem(
		"Phone 1",
		[]string{"Electronics", "Phones"},
		"desc",
		sellerID,
		100,
		time.Now().Add(-time.Hour),
		time.Now().Add(time.Hour),
	)
	if err != nil {
		t.Fatalf("CreateItem failed: %v", err)
	}
	_, err = service.CreateItem(
		"Laptop 1",
		[]string{"Electronics", "Laptops"},
		"desc",
		sellerID,
		100,
		time.Now().Add(-time.Hour),
		time.Now().Add(time.Hour),
	)
	if err != nil {
		t.Fatalf("CreateItem failed: %v", err)
	}

	items, err := service.BrowseCategory("Phones")
	if err != nil {
		t.Fatalf("BrowseCategory failed: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item in Phones subtree, got %d", len(items))
	}
	if items[0].Name != "Phone 1" {
		t.Fatalf("expected Phone 1, got %s", items[0].Name)
	}
}

func TestBrowseCategoryMissing(t *testing.T) {
	store := NewAuctionStore()
	service := NewAuctionService(store)

	_, err := service.BrowseCategory("Unknown")
	if !errors.Is(err, ErrCategoryNotFound) {
		t.Fatalf("expected ErrCategoryNotFound, got %v", err)
	}
}

func mustSeedUser(t *testing.T, store *AuctionStore, name string, balance float64) models.UserID {
	t.Helper()
	userID := store.generateUserID()
	store.mu.Lock()
	store.usersByID[userID] = &models.User{
		ID:      userID,
		Name:    name,
		Balance: balance,
	}
	store.mu.Unlock()
	return userID
}
