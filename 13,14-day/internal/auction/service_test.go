package auction

import (
	"errors"
	"math"
	"sync"
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

func TestEndAuctionReturnsWinnerAndMarksEnded(t *testing.T) {
	store := NewAuctionStore()
	service := NewAuctionService(store)

	userA, itemID := seedUserAndItem(t, store, 1_000, 100, 2)
	userB := models.UserID(3)
	store.mu.Lock()
	store.usersByID[userB] = &models.User{ID: userB, Name: "user-b", Balance: 1_000}
	store.mu.Unlock()

	if _, err := service.PlaceBid(userA, itemID, 150); err != nil {
		t.Fatalf("first bid failed: %v", err)
	}
	secondBidID, err := service.PlaceBid(userB, itemID, 200)
	if err != nil {
		t.Fatalf("second bid failed: %v", err)
	}

	winner, err := service.EndAuction(itemID)
	if err != nil {
		t.Fatalf("EndAuction failed: %v", err)
	}
	if winner == nil {
		t.Fatal("expected winner, got nil")
	}
	if winner.ID != secondBidID {
		t.Fatalf("expected winner bidID %d, got %d", secondBidID, winner.ID)
	}

	store.mu.RLock()
	defer store.mu.RUnlock()
	if store.itemsByID[itemID].Status != models.ItemStatusEnded {
		t.Fatalf("expected status ended, got %s", store.itemsByID[itemID].Status)
	}
}

func TestEndAuctionNoBids(t *testing.T) {
	store := NewAuctionStore()
	service := NewAuctionService(store)

	_, itemID := seedUserAndItem(t, store, 1_000, 100, 2)
	winner, err := service.EndAuction(itemID)
	if err != nil {
		t.Fatalf("EndAuction failed: %v", err)
	}
	if winner != nil {
		t.Fatalf("expected nil winner, got bid %d", winner.ID)
	}
}

// TestCreateUserValidationErrorType exists because user/item input errors were previously
// reported as bid errors, which made API responses and diagnostics misleading.
func TestCreateUserValidationErrorType(t *testing.T) {
	store := NewAuctionStore()
	service := NewAuctionService(store)

	_, err := service.CreateUser("   ", 100)
	if !errors.Is(err, ErrInvalidUserInput) {
		t.Fatalf("expected ErrInvalidUserInput, got %v", err)
	}
}

// TestCreateItemNormalizesCategoryPath exists to prevent empty/whitespace category leafs
// from being stored under an invalid category key.
func TestCreateItemNormalizesCategoryPath(t *testing.T) {
	store := NewAuctionStore()
	service := NewAuctionService(store)

	sellerID := models.UserID(10)
	now := time.Now().UTC()
	store.mu.Lock()
	store.usersByID[sellerID] = &models.User{
		ID:      sellerID,
		Name:    "seller",
		Balance: 1_000,
	}
	store.mu.Unlock()

	itemID, err := service.CreateItem(
		"item",
		[]string{"  Electronics  ", "  ", " Phones "},
		"desc",
		sellerID,
		100,
		now.Add(-time.Hour),
		now.Add(time.Hour),
	)
	if err != nil {
		t.Fatalf("CreateItem failed: %v", err)
	}

	store.mu.RLock()
	defer store.mu.RUnlock()
	item := store.itemsByID[itemID]
	if item.Category != "Phones" {
		t.Fatalf("expected normalized leaf category Phones, got %q", item.Category)
	}
	if _, ok := store.itemIDsByCategory[""]; ok {
		t.Fatalf("unexpected empty category index entry")
	}
}

// TestRetractDoesNotDuplicateBidHistory exists because retract previously pushed the same
// bid ID into history again, creating inconsistent bid-history API output.
func TestRetractDoesNotDuplicateBidHistory(t *testing.T) {
	store := NewAuctionStore()
	service := NewAuctionService(store)
	userID, itemID := seedUserAndItem(t, store, 1_000, 100, 2)

	bidID, err := service.PlaceBid(userID, itemID, 150)
	if err != nil {
		t.Fatalf("place bid failed: %v", err)
	}

	if _, err := service.RetractLastBid(userID, itemID); err != nil {
		t.Fatalf("retract failed: %v", err)
	}

	item, err := service.GetItem(itemID)
	if err != nil {
		t.Fatalf("GetItem failed: %v", err)
	}
	if len(item.BidHistory) != 1 {
		t.Fatalf("expected bid history length 1, got %d", len(item.BidHistory))
	}
	if item.BidHistory[0] != bidID {
		t.Fatalf("expected history to keep original bid id %d, got %d", bidID, item.BidHistory[0])
	}
}

func TestConcurrentBiddingSingleItem100Goroutines(t *testing.T) {
	store := NewAuctionStore()
	service := NewAuctionService(store)

	sellerID := models.UserID(2)
	itemID := models.ItemID(99)
	now := time.Now().UTC()

	store.mu.Lock()
	store.itemsByID[itemID] = &models.Item{
		ID:         itemID,
		Name:       "concurrent-item",
		SellerID:   sellerID,
		StartPrice: 100,
		StartTime:  now.Add(-time.Hour),
		EndTime:    now.Add(time.Hour),
		Status:     models.ItemStatusActive,
	}
	for i := 1; i <= 100; i++ {
		uid := models.UserID(i + 10)
		store.usersByID[uid] = &models.User{
			ID:      uid,
			Name:    "bidder",
			Balance: 1_000_000,
		}
	}
	store.mu.Unlock()

	var wg sync.WaitGroup
	wg.Add(100)

	for i := 1; i <= 100; i++ {
		i := i
		go func() {
			defer wg.Done()
			uid := models.UserID(i + 10)
			amount := float64(100 + i)
			_, _ = service.PlaceBid(uid, itemID, amount)
		}()
	}

	wg.Wait()

	winner, err := service.EndAuction(itemID)
	if err != nil {
		t.Fatalf("EndAuction failed: %v", err)
	}
	if winner == nil {
		t.Fatal("expected winner after concurrent bidding, got nil")
	}
	if winner.Amount != 200 {
		t.Fatalf("expected highest winning amount 200, got %v", winner.Amount)
	}
}

func TestRetractLastBidTwiceFails(t *testing.T) {
	store := NewAuctionStore()
	service := NewAuctionService(store)

	userID, itemID := seedUserAndItem(t, store, 1_000, 100, 2)
	if _, err := service.PlaceBid(userID, itemID, 150); err != nil {
		t.Fatalf("place bid failed: %v", err)
	}

	if _, err := service.RetractLastBid(userID, itemID); err != nil {
		t.Fatalf("first retract failed: %v", err)
	}

	_, err := service.RetractLastBid(userID, itemID)
	if !errors.Is(err, ErrNoBidToRetract) {
		t.Fatalf("expected ErrNoBidToRetract on second retract, got %v", err)
	}
}

func TestRetractOtherUserBidFails(t *testing.T) {
	store := NewAuctionStore()
	service := NewAuctionService(store)

	userA, itemID := seedUserAndItem(t, store, 1_000, 100, 2)
	userB := models.UserID(777)
	store.mu.Lock()
	store.usersByID[userB] = &models.User{ID: userB, Name: "user-b", Balance: 1_000}
	store.mu.Unlock()

	if _, err := service.PlaceBid(userA, itemID, 150); err != nil {
		t.Fatalf("place bid by userA failed: %v", err)
	}

	_, err := service.RetractLastBid(userB, itemID)
	if !errors.Is(err, ErrNoBidToRetract) {
		t.Fatalf("expected ErrNoBidToRetract for user without bids, got %v", err)
	}
}

func TestRetractWhenAuctionEndedFails(t *testing.T) {
	store := NewAuctionStore()
	service := NewAuctionService(store)

	userID, itemID := seedUserAndItem(t, store, 1_000, 100, 2)
	if _, err := service.PlaceBid(userID, itemID, 150); err != nil {
		t.Fatalf("place bid failed: %v", err)
	}
	if _, err := service.EndAuction(itemID); err != nil {
		t.Fatalf("end auction failed: %v", err)
	}

	_, err := service.RetractLastBid(userID, itemID)
	if !errors.Is(err, ErrAuctionNotActive) {
		t.Fatalf("expected ErrAuctionNotActive after auction end, got %v", err)
	}
}

func TestPlaceBidRejectsNaNAndInf(t *testing.T) {
	store := NewAuctionStore()
	service := NewAuctionService(store)
	userID, itemID := seedUserAndItem(t, store, 1_000, 100, 2)

	_, err := service.PlaceBid(userID, itemID, math.NaN())
	if !errors.Is(err, ErrInvalidBidAmount) {
		t.Fatalf("expected ErrInvalidBidAmount for NaN, got %v", err)
	}

	_, err = service.PlaceBid(userID, itemID, math.Inf(1))
	if !errors.Is(err, ErrInvalidBidAmount) {
		t.Fatalf("expected ErrInvalidBidAmount for Inf, got %v", err)
	}
}

func TestPlaceBidRejectsCancelledItem(t *testing.T) {
	store := NewAuctionStore()
	service := NewAuctionService(store)
	userID, itemID := seedUserAndItem(t, store, 1_000, 100, 2)
	store.mu.Lock()
	store.itemsByID[itemID].Status = models.ItemStatusCancelled
	store.mu.Unlock()

	_, err := service.PlaceBid(userID, itemID, 150)
	if !errors.Is(err, ErrAuctionNotActive) {
		t.Fatalf("expected ErrAuctionNotActive, got %v", err)
	}
}

func TestPlaceBidRejectsOutsideTimeWindow(t *testing.T) {
	store := NewAuctionStore()
	service := NewAuctionService(store)
	userID := models.UserID(1)
	itemID := models.ItemID(10)
	now := time.Now().UTC()

	store.mu.Lock()
	store.usersByID[userID] = &models.User{ID: userID, Name: "u", Balance: 1_000}
	store.itemsByID[itemID] = &models.Item{
		ID:         itemID,
		Name:       "item",
		SellerID:   models.UserID(2),
		StartPrice: 100,
		StartTime:  now.Add(time.Hour),
		EndTime:    now.Add(2 * time.Hour),
		Status:     models.ItemStatusActive,
	}
	store.mu.Unlock()

	_, err := service.PlaceBid(userID, itemID, 150)
	if !errors.Is(err, ErrAuctionNotActive) {
		t.Fatalf("expected ErrAuctionNotActive before start, got %v", err)
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
