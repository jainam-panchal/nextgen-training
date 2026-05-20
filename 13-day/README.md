# Day 13 - Real-Time Auction Platform (Part 1)

## Current Status
Core Day 13 requirements are implemented:
- Data models: `User`, `Item`, `Bid` with typed IDs.
- Data structures:
  - Max-Heap for active bids per item.
  - HashMaps for users/items/bids.
  - Tree for category hierarchy (`internal/ds/tree`).
  - Linked list for per-item bid history.
  - Stack for per-user-per-item undo/retract.
- Core logic:
  - `PlaceBid`
  - `RetractLastBid`
  - `EndAuction`
  - `BrowseCategory`
- Concurrency:
  - Fine-grained lock per item (`itemLock`) + store-wide RW lock (`store.mu`).
  - 100-goroutine concurrent bidding test on same item.

## Architecture (Simplified)
`AuctionStore` keeps authoritative state:
- `usersByID`, `itemsByID`, `bidsByID`
- `itemBidHeaps[itemID]`
- `itemBidHistory[itemID]`
- `userBidStacks[userID][itemID]`
- `categoryTree` + `itemIDsByCategory`
- `itemLiveQueues[itemID]`

Locking pattern:
- Always lock in order: `itemLock(itemID)` -> `store.mu`
- `PlaceBid`, `RetractLastBid`, `EndAuction` follow this order.

## Recent Changes (Last Set)
1. Moved category DS to generic package:
   - `internal/ds/tree`
2. Added category browse support:
   - category tree indexing + subtree lookup
3. Added `EndAuction(itemID)`:
   - marks item ended, computes winner from heap, emits ended event
4. Added stress test:
   - `TestConcurrentBiddingSingleItem100Goroutines`
5. Added retract edge-case tests:
   - double retract fails
   - wrong user retract fails
   - retract after auction end fails

## Tests Run
Commands executed:
```bash
go test ./...
go test -race -count=5 ./...
```

Latest observed result:
```text
ok  	realtime-auction/internal/auction
ok  	realtime-auction/internal/ds/heap
ok  	realtime-auction/internal/ds/linkedlist
ok  	realtime-auction/internal/ds/stack
ok  	realtime-auction/internal/ds/tree
?   	realtime-auction/internal/models	[no test files]
```

## Short Note (What Was Tested)
Tested happy paths and failure paths for bidding, retracting, ending auctions, and category browsing. Also tested concurrent writes (100 goroutines on one item) and repeated race runs (`-race -count=5`) to validate lock safety and state consistency.

## Remaining Optional Polish
- Add explicit tests for bid rejection:
  - before start time
  - after end time
  - cancelled item status
- Add minimal API/handler layer in Day 14.
