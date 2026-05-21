package auction

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"realtime-auction/internal/models"
)

func BenchmarkCore_PlaceBid_NoContention(b *testing.B) {
	tiers := []int{100, 1000, 10000}

	for _, users := range tiers {
		b.Run(fmt.Sprintf("users_%d", users), func(b *testing.B) {
			_, service, bidderIDs, itemIDs := setupCoreBench(users)
			seq := make([]atomic.Int64, len(itemIDs))
			for i := range seq {
				seq[i].Store(1000)
			}

			var attempts atomic.Int64
			var success atomic.Int64

			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				idx := int(attempts.Load()) % len(bidderIDs)
				userID := bidderIDs[idx]
				itemID := itemIDs[idx]
				amount := float64(seq[idx].Add(1))

				attempts.Add(1)
				if _, err := service.PlaceBid(userID, itemID, amount); err == nil {
					success.Add(1)
				}
			}

			if attempts.Load() == 0 || success.Load() != attempts.Load() {
				b.Fatalf("expected all no-contention bids to succeed; attempts=%d success=%d", attempts.Load(), success.Load())
			}
		})
	}
}

func BenchmarkCore_PlaceBid_Contention(b *testing.B) {
	tiers := []int{100, 1000, 10000}

	for _, users := range tiers {
		b.Run(fmt.Sprintf("users_%d", users), func(b *testing.B) {
			_, service, bidderIDs, itemIDs := setupCoreBench(users)
			itemID := itemIDs[0]

			var seq atomic.Int64
			seq.Store(1000)

			var attempts atomic.Int64
			var success atomic.Int64

			b.ReportAllocs()
			b.ResetTimer()

			b.RunParallel(func(pb *testing.PB) {
				var i int
				for pb.Next() {
					userID := bidderIDs[i%len(bidderIDs)]
					amount := float64(seq.Add(1))

					attempts.Add(1)
					if _, err := service.PlaceBid(userID, itemID, amount); err == nil {
						success.Add(1)
					}
					i++
				}
			})

			if attempts.Load() == 0 || success.Load() == 0 {
				b.Fatalf("expected successful bids under contention; attempts=%d success=%d", attempts.Load(), success.Load())
			}
		})
	}
}

func BenchmarkCore_RetractLastBid(b *testing.B) {
	tiers := []int{100, 1000, 10000}

	for _, users := range tiers {
		b.Run(fmt.Sprintf("users_%d", users), func(b *testing.B) {
			store, service, bidderIDs, itemIDs := setupCoreBench(users)
			itemID := itemIDs[0]
			var seq atomic.Int64
			seq.Store(10_000)

			var attempts atomic.Int64
			var success atomic.Int64

			b.ReportAllocs()
			for b.Loop() {
				userID := bidderIDs[int(attempts.Load())%len(bidderIDs)]

				b.StopTimer()
				amount := float64(seq.Add(1))
				if _, err := service.PlaceBid(userID, itemID, amount); err != nil {
					b.Fatalf("setup place bid failed: %v", err)
				}
				b.StartTimer()

				attempts.Add(1)
				if _, err := service.RetractLastBid(userID, itemID); err == nil {
					success.Add(1)
				}
			}

			if success.Load() != attempts.Load() {
				b.Fatalf("expected all retracts to succeed; attempts=%d success=%d", attempts.Load(), success.Load())
			}

			_ = store
		})
	}
}

func BenchmarkCore_EndAuction(b *testing.B) {
	tiers := []int{100, 1000, 10000}

	for _, bidCount := range tiers {
		b.Run(fmt.Sprintf("bid_count_%d", bidCount), func(b *testing.B) {
			store, service, bidderIDs, itemIDs := setupCoreBench(max(2, min(bidCount, 1000)))
			itemID := itemIDs[0]
			var seq atomic.Int64
			seq.Store(1000)

			for i := 0; i < bidCount; i++ {
				userID := bidderIDs[i%len(bidderIDs)]
				amount := float64(seq.Add(1))
				if _, err := service.PlaceBid(userID, itemID, amount); err != nil {
					b.Fatalf("setup place bid failed: %v", err)
				}
			}

			var attempts, success atomic.Int64
			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				attempts.Add(1)
				if winner, err := service.EndAuction(itemID); err == nil && winner != nil {
					success.Add(1)
				}

				b.StopTimer()
				lock := store.getItemLock(itemID)
				lock.Lock()
				store.mu.Lock()
				item := store.itemsByID[itemID]
				item.Status = models.ItemStatusActive
				store.mu.Unlock()
				lock.Unlock()
				b.StartTimer()
			}

			if success.Load() != attempts.Load() {
				b.Fatalf("expected all EndAuction calls to produce winner; attempts=%d success=%d", attempts.Load(), success.Load())
			}
		})
	}
}

func BenchmarkCore_BrowseCategory_Subtree(b *testing.B) {
	tiers := []int{100, 1000, 10000}

	for _, items := range tiers {
		b.Run(fmt.Sprintf("items_%d", items), func(b *testing.B) {
			store := NewAuctionStore()
			service := NewAuctionService(store)
			now := time.Now().UTC()
			start := now.Add(-time.Hour)
			end := now.Add(time.Hour)

			sellerID, err := service.CreateUser("seller", 1_000_000)
			if err != nil {
				b.Fatalf("create seller failed: %v", err)
			}

			for i := 0; i < items; i++ {
				_, err := service.CreateItem(
					fmt.Sprintf("item-%d", i),
					[]string{"Electronics", "Phones", "Smartphones"},
					"bench",
					sellerID,
					100,
					start,
					end,
				)
				if err != nil {
					b.Fatalf("setup create item failed: %v", err)
				}
			}

			var count int64
			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				out, err := service.BrowseCategory("Electronics")
				if err != nil {
					b.Fatalf("browse failed: %v", err)
				}
				if len(out) == 0 {
					b.Fatalf("expected non-empty browse output")
				}
				atomic.AddInt64(&count, int64(len(out)))
			}

			if atomic.LoadInt64(&count) == 0 {
				b.Fatalf("expected browse to return items")
			}
		})
	}
}

func BenchmarkCore_LivePublish_Fanout(b *testing.B) {
	watcherTiers := []int{10, 100, 500}

	for _, watchers := range watcherTiers {
		b.Run(fmt.Sprintf("watchers_%d", watchers), func(b *testing.B) {
			_, service, bidderIDs, itemIDs := setupCoreBench(100)
			itemID := itemIDs[0]
			var seq atomic.Int64
			seq.Store(1000)

			watcherChans := make([]chan BidEvent, 0, watchers)
			done := make(chan struct{})
			for i := 0; i < watchers; i++ {
				ch, err := service.SubscribeItem(itemID)
				if err != nil {
					b.Fatalf("subscribe failed: %v", err)
				}
				watcherChans = append(watcherChans, ch)
				go func(c chan BidEvent) {
					for {
						select {
						case <-done:
							return
						case <-c:
						}
					}
				}(ch)
			}
			b.Cleanup(func() {
				close(done)
				for _, ch := range watcherChans {
					service.UnsubscribeItem(itemID, ch)
				}
			})

			var attempts atomic.Int64
			var success atomic.Int64

			b.ReportAllocs()
			b.ResetTimer()

			b.RunParallel(func(pb *testing.PB) {
				var i int
				for pb.Next() {
					userID := bidderIDs[i%len(bidderIDs)]
					amount := float64(seq.Add(1))
					attempts.Add(1)
					if _, err := service.PlaceBid(userID, itemID, amount); err == nil {
						success.Add(1)
					}
					i++
				}
			})

			if attempts.Load() == 0 || success.Load() == 0 {
				b.Fatalf("expected successful bids while fanout active; attempts=%d success=%d", attempts.Load(), success.Load())
			}
		})
	}
}

func setupCoreBench(users int) (*AuctionStore, *AuctionService, []models.UserID, []models.ItemID) {
	store := NewAuctionStore()
	service := NewAuctionService(store)
	now := time.Now().UTC()
	start := now.Add(-time.Hour)
	end := now.Add(time.Hour)

	sellerID, err := service.CreateUser("seller", 1_000_000_000_000)
	if err != nil {
		panic(err)
	}

	bidders := make([]models.UserID, 0, users)
	for i := 0; i < users; i++ {
		userID, userErr := service.CreateUser(fmt.Sprintf("bidder-%d", i), 1_000_000_000_000)
		if userErr != nil {
			panic(userErr)
		}
		bidders = append(bidders, userID)
	}

	itemIDs := make([]models.ItemID, 0, users)
	for i := 0; i < users; i++ {
		itemID, itemErr := service.CreateItem(
			fmt.Sprintf("item-%d", i),
			[]string{"Electronics", "Phones"},
			"bench",
			sellerID,
			100,
			start,
			end,
		)
		if itemErr != nil {
			panic(itemErr)
		}
		itemIDs = append(itemIDs, itemID)
	}

	return store, service, bidders, itemIDs
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
