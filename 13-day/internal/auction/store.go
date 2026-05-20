package auction

import (
	"realtime-auction/internal/ds/heap"
	"realtime-auction/internal/ds/linkedlist"
	"realtime-auction/internal/ds/stack"
	"realtime-auction/internal/models"
	"sync"
	"sync/atomic"
)

type AuctionStore struct {
	mu sync.RWMutex

	usersByID map[models.UserID]*models.User
	itemsByID map[models.ItemID]*models.Item
	bidsByID  map[models.BidID]*models.Bid

	itemBidHeaps   map[models.ItemID]*heap.MaxHeap[*models.Bid]
	itemBidHistory map[models.ItemID]*linkedlist.LinkedList[models.BidID]
	itemLiveQueues map[models.ItemID]chan *models

	userBidStacks map[models.UserID]*stack.Stack[models.BidID]

	itemLocksMu sync.Mutex
	itemLocks   map[models.ItemID]*sync.Mutex

	nextUserID atomic.Int64
	nextItemID atomic.Int64
	nextBidID  atomic.Int64
}

func NewAuctionStore() *AuctionStore {
	return &AuctionStore{
		usersByID:      make(map[models.UserID]*models.User),
		itemsByID:      make(map[models.ItemID]*models.Item),
		bidsByID:       make(map[models.BidID]*models.Bid),
		itemBidHeaps:   make(map[models.ItemID]*heap.MaxHeap[*models.Bid]),
		itemBidHistory: make(map[models.ItemID]*linkedlist.LinkedList[models.BidID]),
		userBidStacks:  make(map[models.UserID]*stack.Stack[models.BidID]),
		itemLocks:      make(map[models.ItemID]*sync.Mutex),
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

func (s *AuctionStore) getOrCreateItemHeap(itemID models.ItemID) *heap.MaxHeap[*models.Bid] {

	s.mu.Lock()
	defer s.mu.Unlock()

	if h, exists := s.itemBidHeaps[itemID]; exists {
		return h
	}

	h := heap.NewMaxHeap(func(a, b *models.Bid) bool {
		return a.Amount < b.Amount
	})
	s.itemBidHeaps[itemID] = h
	return h
}

func (s *AuctionStore) getOrCreateItemBidHistory(itemID models.ItemID) *linkedlist.LinkedList[models.BidID] {

	s.mu.Lock()
	defer s.mu.Unlock()

	if h, exists := s.itemBidHistory[itemID]; exists {
		return h
	}

	h := linkedlist.NewLinkedList[models.BidID]()
	s.itemBidHistory[itemID] = h
	return h
}

func (s *AuctionStore) getOrCreateUserBidStack(userID models.UserID) *stack.Stack[models.BidID] {

	s.mu.Lock()
	defer s.mu.Unlock()

	if h, exists := s.userBidStacks[userID]; exists {
		return h
	}

	h := stack.NewStack[models.BidID]()
	s.userBidStacks[userID] = h
	return h
}
