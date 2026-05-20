package auction

import (
	"errors"
	"testing"
	"time"

	"realtime-auction/internal/models"
)

// TestPlaceBidSuccess verifies the happy path: valid bidder, active item, and higher bid updates winner.
func TestPlaceBidSuccess(t *testing.T) {
	store := NewAuctionStore()
	service := NewAuctionService(store)

	userID, itemID := seedUserAndItem(t, store, 1_000, 100, 2)

	bidID, err := service.PlaceBid(userID, itemID, 150)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if bidID == 0 {
		t.Fatal("expected non-zero bidID")
	}

	store.mu.RLock()
	defer store.mu.RUnlock()

	item := store.itemsByID[itemID]
	if item.CurrentBid == nil {
		t.Fatal("expected current bid to be set")
	}
	if item.CurrentBid.ID != bidID {
		t.Fatalf("expected current bid ID %d, got %d", bidID, item.CurrentBid.ID)
	}
}

// TestPlaceBidRejectsLowerAmount ensures bids not higher than current/start price are rejected.
func TestPlaceBidRejectsLowerAmount(t *testing.T) {
	store := NewAuctionStore()
	service := NewAuctionService(store)
	userID, itemID := seedUserAndItem(t, store, 1_000, 100, 2)

	if _, err := service.PlaceBid(userID, itemID, 150); err != nil {
		t.Fatalf("seed bid failed: %v", err)
	}

	_, err := service.PlaceBid(userID, itemID, 140)
	if !errors.Is(err, ErrInvalidBidAmount) {
		t.Fatalf("expected ErrInvalidBidAmount, got %v", err)
	}
}

// TestPlaceBidRejectsSeller verifies seller cannot place bid on own item.
func TestPlaceBidRejectsSeller(t *testing.T) {
	store := NewAuctionStore()
	service := NewAuctionService(store)
	sellerID := models.UserID(10)
	itemID := models.ItemID(20)

	now := time.Now().UTC()
	store.mu.Lock()
	store.usersByID[sellerID] = &models.User{
		ID:      sellerID,
		Name:    "seller",
		Balance: 1_000,
	}
	store.itemsByID[itemID] = &models.Item{
		ID:         itemID,
		Name:       "item",
		SellerID:   sellerID,
		StartPrice: 100,
		StartTime:  now.Add(-time.Hour),
		EndTime:    now.Add(time.Hour),
		Status:     models.ItemStatusActive,
	}
	store.mu.Unlock()

	_, err := service.PlaceBid(sellerID, itemID, 200)
	if !errors.Is(err, ErrSellerCannotBid) {
		t.Fatalf("expected ErrSellerCannotBid, got %v", err)
	}
}

// TestRetractLastBidSuccess verifies retract marks bid inactive and rolls winner back to previous active bid.
func TestRetractLastBidSuccess(t *testing.T) {
	store := NewAuctionStore()
	service := NewAuctionService(store)

	userA, itemID := seedUserAndItem(t, store, 1_000, 100, 2)
	userB := models.UserID(3)
	store.mu.Lock()
	store.usersByID[userB] = &models.User{ID: userB, Name: "user-b", Balance: 1_000}
	store.mu.Unlock()

	firstBidID, err := service.PlaceBid(userA, itemID, 150)
	if err != nil {
		t.Fatalf("first bid failed: %v", err)
	}
	secondBidID, err := service.PlaceBid(userB, itemID, 200)
	if err != nil {
		t.Fatalf("second bid failed: %v", err)
	}
	if firstBidID == secondBidID {
		t.Fatal("expected distinct bid IDs")
	}

	retractedID, err := service.RetractLastBid(userB, itemID)
	if err != nil {
		t.Fatalf("retract failed: %v", err)
	}
	if retractedID != secondBidID {
		t.Fatalf("expected retracted bid %d, got %d", secondBidID, retractedID)
	}

	store.mu.RLock()
	defer store.mu.RUnlock()

	item := store.itemsByID[itemID]
	if item.CurrentBid == nil {
		t.Fatal("expected previous winning bid after retract")
	}
	if item.CurrentBid.ID != firstBidID {
		t.Fatalf("expected current bid to roll back to %d, got %d", firstBidID, item.CurrentBid.ID)
	}
	if !store.bidsByID[secondBidID].IsRetracted {
		t.Fatal("expected retracted bid to be marked inactive")
	}
}

// TestRetractLastBidEmptyStack ensures retract fails when user has no bid for that item.
func TestRetractLastBidEmptyStack(t *testing.T) {
	store := NewAuctionStore()
	service := NewAuctionService(store)

	userID, itemID := seedUserAndItem(t, store, 1_000, 100, 2)
	_, err := service.RetractLastBid(userID, itemID)
	if !errors.Is(err, ErrNoBidToRetract) {
		t.Fatalf("expected ErrNoBidToRetract, got %v", err)
	}
}

// TestRetractIsItemScoped verifies retract for one item does not pop stack entries of another item.
func TestRetractIsItemScoped(t *testing.T) {
	store := NewAuctionStore()
	service := NewAuctionService(store)
	userID := models.UserID(1)
	otherUser := models.UserID(2)
	itemA := models.ItemID(11)
	itemB := models.ItemID(12)

	now := time.Now().UTC()
	store.mu.Lock()
	store.usersByID[userID] = &models.User{ID: userID, Name: "u1", Balance: 5_000}
	store.usersByID[otherUser] = &models.User{ID: otherUser, Name: "u2", Balance: 5_000}
	store.itemsByID[itemA] = &models.Item{
		ID:         itemA,
		Name:       "A",
		SellerID:   otherUser,
		StartPrice: 100,
		StartTime:  now.Add(-time.Hour),
		EndTime:    now.Add(time.Hour),
		Status:     models.ItemStatusActive,
	}
	store.itemsByID[itemB] = &models.Item{
		ID:         itemB,
		Name:       "B",
		SellerID:   otherUser,
		StartPrice: 100,
		StartTime:  now.Add(-time.Hour),
		EndTime:    now.Add(time.Hour),
		Status:     models.ItemStatusActive,
	}
	store.mu.Unlock()

	bidA, err := service.PlaceBid(userID, itemA, 150)
	if err != nil {
		t.Fatalf("bid on itemA failed: %v", err)
	}
	bidB, err := service.PlaceBid(userID, itemB, 160)
	if err != nil {
		t.Fatalf("bid on itemB failed: %v", err)
	}

	// Retracting on itemA must retract bidA only, not bidB.
	got, err := service.RetractLastBid(userID, itemA)
	if err != nil {
		t.Fatalf("retract on itemA failed: %v", err)
	}
	if got != bidA {
		t.Fatalf("expected retracted bid %d for itemA, got %d", bidA, got)
	}

	store.mu.RLock()
	defer store.mu.RUnlock()
	if store.bidsByID[bidB].IsRetracted {
		t.Fatalf("expected bid %d for itemB to remain active", bidB)
	}
}

func seedUserAndItem(t *testing.T, store *AuctionStore, balance float64, startPrice float64, sellerID models.UserID) (models.UserID, models.ItemID) {
	t.Helper()

	userID := models.UserID(1)
	itemID := models.ItemID(10)
	now := time.Now().UTC()

	store.mu.Lock()
	store.usersByID[userID] = &models.User{
		ID:      userID,
		Name:    "user-1",
		Balance: balance,
	}
	store.itemsByID[itemID] = &models.Item{
		ID:         itemID,
		Name:       "item-1",
		SellerID:   sellerID,
		StartPrice: startPrice,
		StartTime:  now.Add(-time.Hour),
		EndTime:    now.Add(time.Hour),
		Status:     models.ItemStatusActive,
	}
	store.mu.Unlock()

	return userID, itemID
}
