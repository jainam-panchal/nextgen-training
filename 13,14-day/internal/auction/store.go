package auction

import (
	"realtime-auction/internal/ds/heap"
	"realtime-auction/internal/ds/linkedlist"
	"realtime-auction/internal/ds/stack"
	"realtime-auction/internal/ds/tree"
	"realtime-auction/internal/models"
	"sync"
	"sync/atomic"
)

const (
	defaultWatcherChannelSize = 64
)

type AuctionStore struct {
	mu sync.RWMutex

	categoryTree      *tree.Tree[string]
	itemIDsByCategory map[string]map[models.ItemID]struct{}

	usersByID map[models.UserID]*models.User
	itemsByID map[models.ItemID]*models.Item
	bidsByID  map[models.BidID]*models.Bid

	itemBidHeaps     map[models.ItemID]*heap.MaxHeap[*models.Bid]
	itemBidHistory   map[models.ItemID]*linkedlist.LinkedList[models.BidID]
	watchersByItemID map[models.ItemID]map[chan BidEvent]struct{}

	userBidStacks map[models.UserID]map[models.ItemID]*stack.Stack[models.BidID]

	itemLocksMu sync.Mutex
	itemLocks   map[models.ItemID]*sync.Mutex

	nextUserID atomic.Int64
	nextItemID atomic.Int64
	nextBidID  atomic.Int64
}

func NewAuctionStore() *AuctionStore {
	return &AuctionStore{
		categoryTree: tree.NewTree("ALL"),
		itemIDsByCategory: map[string]map[models.ItemID]struct{}{
			"ALL": make(map[models.ItemID]struct{}),
		},
		usersByID:        make(map[models.UserID]*models.User),
		itemsByID:        make(map[models.ItemID]*models.Item),
		bidsByID:         make(map[models.BidID]*models.Bid),
		itemBidHeaps:     make(map[models.ItemID]*heap.MaxHeap[*models.Bid]),
		itemBidHistory:   make(map[models.ItemID]*linkedlist.LinkedList[models.BidID]),
		userBidStacks:    make(map[models.UserID]map[models.ItemID]*stack.Stack[models.BidID]),
		itemLocks:        make(map[models.ItemID]*sync.Mutex),
		watchersByItemID: make(map[models.ItemID]map[chan BidEvent]struct{}),
	}
}

func (s *AuctionStore) generateUserID() models.UserID {
	return models.UserID(s.nextUserID.Add(1))
}

func (s *AuctionStore) generateItemID() models.ItemID {
	return models.ItemID(s.nextItemID.Add(1))
}

func (s *AuctionStore) generateBidID() models.BidID {
	return models.BidID(s.nextBidID.Add(1))
}

func (s *AuctionStore) getItemLock(itemID models.ItemID) *sync.Mutex {

	s.itemLocksMu.Lock()
	defer s.itemLocksMu.Unlock()

	if lock, exists := s.itemLocks[itemID]; exists {
		return lock
	}

	lock := &sync.Mutex{}
	s.itemLocks[itemID] = lock
	return lock
}

// The following helpers require s.mu to already be held by caller.

func (s *AuctionStore) getOrCreateItemHeapLocked(itemID models.ItemID) *heap.MaxHeap[*models.Bid] {
	if h, exists := s.itemBidHeaps[itemID]; exists {
		return h
	}

	h := heap.NewMaxHeap(func(a, b *models.Bid) bool {
		return a.Amount < b.Amount
	})
	s.itemBidHeaps[itemID] = h
	return h
}

func (s *AuctionStore) getOrCreateItemBidHistoryLocked(itemID models.ItemID) *linkedlist.LinkedList[models.BidID] {
	if h, exists := s.itemBidHistory[itemID]; exists {
		return h
	}

	h := linkedlist.NewLinkedList[models.BidID]()
	s.itemBidHistory[itemID] = h
	return h
}

func (s *AuctionStore) getOrCreateUserItemBidStackLocked(userID models.UserID, itemID models.ItemID) *stack.Stack[models.BidID] {
	itemStacks, exists := s.userBidStacks[userID]
	if !exists {
		itemStacks = make(map[models.ItemID]*stack.Stack[models.BidID])
		s.userBidStacks[userID] = itemStacks
	}

	if h, exists := itemStacks[itemID]; exists {
		return h
	}

	h := stack.NewStack[models.BidID]()
	itemStacks[itemID] = h
	return h
}

func (s *AuctionStore) ensureCategoryPathLocked(path []string) error {
	if len(path) == 0 {
		return nil
	}
	parent := "ALL"
	for _, category := range path {
		if category == "" {
			continue
		}
		if _, exists := s.categoryTree.Get(category); !exists {
			if err := s.categoryTree.Add(parent, category); err != nil && err != tree.ErrChildExists {
				return err
			}
		}
		parent = category
	}
	return nil
}

func (s *AuctionStore) addItemToCategoryLocked(itemID models.ItemID, category string) {
	set, exists := s.itemIDsByCategory[category]
	if !exists {
		set = make(map[models.ItemID]struct{})
		s.itemIDsByCategory[category] = set
	}
	set[itemID] = struct{}{}
}

func (s *AuctionStore) addWatcher(itemID models.ItemID, ch chan BidEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	watchers, exists := s.watchersByItemID[itemID]
	if !exists {
		watchers = make(map[chan BidEvent]struct{})
		s.watchersByItemID[itemID] = watchers
	}
	watchers[ch] = struct{}{}
}

func (s *AuctionStore) removeWatcher(itemID models.ItemID, ch chan BidEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	watchers, exists := s.watchersByItemID[itemID]
	if !exists {
		return
	}
	delete(watchers, ch)
	if len(watchers) == 0 {
		delete(s.watchersByItemID, itemID)
	}
}

func (s *AuctionStore) watcherSnapshot(itemID models.ItemID) []chan BidEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	watchers, exists := s.watchersByItemID[itemID]
	if !exists {
		return nil
	}
	out := make([]chan BidEvent, 0, len(watchers))
	for ch := range watchers {
		out = append(out, ch)
	}
	return out
}

func (s *AuctionStore) publishToWatchers(itemID models.ItemID, event BidEvent) {
	watchers := s.watcherSnapshot(itemID)
	for _, watcher := range watchers {
		select {
		case watcher <- event:
		default:
		}
	}
}
