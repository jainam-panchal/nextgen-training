package auction

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"realtime-auction/internal/models"
)

// TestConcurrentBiddingSimulation50Users5Items simulates:
// - 50 users
// - 5 items
// - each user places 10 bids over ~30 seconds
func TestConcurrentBiddingSimulation50Users5Items(t *testing.T) {
	store := NewAuctionStore()
	service := NewAuctionService(store)

	const (
		userCount    = 50
		itemCount    = 5
		bidsPerUser  = 10
		totalRuntime = 30 * time.Second
	)

	sellerID := models.UserID(1)
	now := time.Now().UTC()

	store.mu.Lock()
	store.usersByID[sellerID] = &models.User{ID: sellerID, Name: "seller", Balance: 1_000_000}
	for i := 0; i < userCount; i++ {
		uid := models.UserID(i + 10)
		store.usersByID[uid] = &models.User{
			ID:      uid,
			Name:    "bidder",
			Balance: 1_000_000,
		}
	}
	itemIDs := make([]models.ItemID, 0, itemCount)
	for i := 0; i < itemCount; i++ {
		itemID := models.ItemID(100 + i)
		itemIDs = append(itemIDs, itemID)
		store.itemsByID[itemID] = &models.Item{
			ID:         itemID,
			Name:       "item",
			Category:   "Electronics",
			SellerID:   sellerID,
			StartPrice: 100,
			StartTime:  now.Add(-time.Hour),
			EndTime:    now.Add(2 * time.Hour),
			Status:     models.ItemStatusActive,
		}
	}
	store.mu.Unlock()

	var seq [itemCount]atomic.Int64
	for i := 0; i < itemCount; i++ {
		seq[i].Store(100)
	}

	interval := totalRuntime / bidsPerUser
	var wg sync.WaitGroup
	wg.Add(userCount)
	start := time.Now()

	for i := 0; i < userCount; i++ {
		uid := models.UserID(i + 10)
		go func(userID models.UserID) {
			defer wg.Done()
			for b := 0; b < bidsPerUser; b++ {
				itemIdx := (int(userID) + b) % itemCount
				itemID := itemIDs[itemIdx]
				amount := float64(seq[itemIdx].Add(1))
				_, _ = service.PlaceBid(userID, itemID, amount)
				time.Sleep(interval)
			}
		}(uid)
	}

	wg.Wait()
	elapsed := time.Since(start)
	if elapsed < 29*time.Second {
		t.Fatalf("simulation ran too short, elapsed=%s", elapsed)
	}

	for i, itemID := range itemIDs {
		winner, err := service.EndAuction(itemID)
		if err != nil {
			t.Fatalf("EndAuction failed for item %d: %v", itemID, err)
		}
		if winner == nil {
			t.Fatalf("expected winner for item %d", itemID)
		}
		expectedMin := float64(seq[i].Load())
		if winner.Amount < expectedMin-2 {
			t.Fatalf("winner amount too low for item %d: got=%v expected around=%v", itemID, winner.Amount, expectedMin)
		}
	}
}
