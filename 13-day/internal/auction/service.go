package auction

import (
	"math"
	"realtime-auction/internal/models"
	"slices"
	"strings"
	"time"
)

type AuctionService struct {
	store *AuctionStore
}

type Stats struct {
	TotalUsers     int     `json:"total_users"`
	TotalItems     int     `json:"total_items"`
	ActiveItems    int     `json:"active_items"`
	EndedItems     int     `json:"ended_items"`
	CancelledItems int     `json:"cancelled_items"`
	TotalBids      int     `json:"total_bids"`
	AverageTopBid  float64 `json:"average_top_bid"`
	RetractedBids  int     `json:"retracted_bids"`
	ActiveWatchers int     `json:"active_watchers"`
}

func NewAuctionService(store *AuctionStore) *AuctionService {
	return &AuctionService{
		store: store,
	}
}

func (s *AuctionService) CreateUser(name string, balance float64) (models.UserID, error) {
	if strings.TrimSpace(name) == "" || balance < 0 {
		return 0, ErrInvalidUserInput
	}

	userID := s.store.generateUserID()
	user := &models.User{
		ID:      userID,
		Name:    name,
		Balance: balance,
	}

	s.store.mu.Lock()
	s.store.usersByID[userID] = user
	s.store.mu.Unlock()

	return userID, nil
}

func (s *AuctionService) CreateItem(
	name string,
	categoryPath []string,
	description string,
	sellerID models.UserID,
	startPrice float64,
	startTime time.Time,
	endTime time.Time,
) (models.ItemID, error) {
	if strings.TrimSpace(name) == "" || startPrice <= 0 || !endTime.After(startTime) {
		return 0, ErrInvalidItemInput
	}
	normalizedCategoryPath := normalizeCategoryPath(categoryPath)
	if len(normalizedCategoryPath) == 0 {
		return 0, ErrCategoryNotFound
	}

	s.store.mu.Lock()
	defer s.store.mu.Unlock()

	if _, ok := s.store.usersByID[sellerID]; !ok {
		return 0, ErrUserNotFound
	}

	if err := s.store.ensureCategoryPathLocked(normalizedCategoryPath); err != nil {
		return 0, err
	}

	itemID := s.store.generateItemID()
	leafCategory := normalizedCategoryPath[len(normalizedCategoryPath)-1]
	item := &models.Item{
		ID:          itemID,
		Name:        name,
		Category:    leafCategory,
		Description: description,
		SellerID:    sellerID,
		StartPrice:  startPrice,
		StartTime:   startTime.UTC(),
		EndTime:     endTime.UTC(),
		Status:      models.ItemStatusActive,
	}

	s.store.itemsByID[itemID] = item
	s.store.addItemToCategoryLocked(itemID, leafCategory)

	return itemID, nil
}

func (s *AuctionService) BrowseCategory(category string) ([]*models.Item, error) {
	category = strings.TrimSpace(category)
	if category == "" {
		return nil, ErrCategoryNotFound
	}

	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	categories, err := s.store.categoryTree.SubtreeValues(category)
	if err != nil {
		return nil, ErrCategoryNotFound
	}

	seen := make(map[models.ItemID]struct{})
	items := make([]*models.Item, 0)

	for _, cat := range categories {
		itemIDs, exists := s.store.itemIDsByCategory[cat]
		if !exists {
			continue
		}
		for itemID := range itemIDs {
			if _, done := seen[itemID]; done {
				continue
			}
			item, ok := s.store.itemsByID[itemID]
			if !ok {
				continue
			}
			seen[itemID] = struct{}{}
			items = append(items, item)
		}
	}

	return items, nil
}

func (s *AuctionService) PlaceBid(userID models.UserID, itemID models.ItemID, amount float64) (models.BidID, error) {
	itemLock := s.store.getItemLock(itemID)
	itemLock.Lock()
	defer itemLock.Unlock()

	now := time.Now().UTC()

	s.store.mu.RLock()

	user, ok := s.store.usersByID[userID]
	if !ok {
		s.store.mu.RUnlock()
		return 0, ErrUserNotFound
	}

	item, ok := s.store.itemsByID[itemID]
	if !ok {
		s.store.mu.RUnlock()
		return 0, ErrItemNotFound
	}

	if math.IsNaN(amount) || math.IsInf(amount, 0) || amount <= 0 {
		s.store.mu.RUnlock()
		return 0, ErrInvalidBidAmount
	}

	if item.Status != models.ItemStatusActive {
		s.store.mu.RUnlock()
		return 0, ErrAuctionNotActive
	}

	if now.Before(item.StartTime) || now.After(item.EndTime) {
		s.store.mu.RUnlock()
		return 0, ErrAuctionNotActive
	}

	if item.SellerID == userID {
		s.store.mu.RUnlock()
		return 0, ErrSellerCannotBid
	}

	if user.Balance < amount {
		s.store.mu.RUnlock()
		return 0, ErrInsufficientBalance
	}

	current := item.StartPrice
	if item.CurrentBid != nil {
		current = item.CurrentBid.Amount
	}
	if amount <= current {
		s.store.mu.RUnlock()
		return 0, ErrInvalidBidAmount
	}

	s.store.mu.RUnlock()

	bidID := s.store.generateBidID()
	bid := &models.Bid{
		ID:        bidID,
		UserID:    userID,
		ItemID:    itemID,
		Amount:    amount,
		Timestamp: now,
	}

	s.store.mu.Lock()

	s.store.bidsByID[bidID] = bid

	h := s.store.getOrCreateItemHeapLocked(itemID)
	h.Push(bid)

	s.store.getOrCreateItemBidHistoryLocked(itemID).AddRear(bidID)
	item.BidHistory = append(item.BidHistory, bidID)

	st := s.store.getOrCreateUserItemBidStackLocked(userID, itemID)
	st.Push(bidID)
	user.ActiveBids = append(user.ActiveBids, bidID)

	item.CurrentBid = bid

	s.store.mu.Unlock()

	s.store.publishToWatchers(itemID, BidEvent{
		ItemID:    itemID,
		BidID:     bidID,
		Action:    BidEventPlaced,
		Timestamp: now,
	})

	return bidID, nil
}

func (s *AuctionService) RetractLastBid(userID models.UserID, itemID models.ItemID) (models.BidID, error) {
	itemLock := s.store.getItemLock(itemID)
	itemLock.Lock()
	defer itemLock.Unlock()

	now := time.Now().UTC()

	// Read validation
	s.store.mu.RLock()

	_, userExists := s.store.usersByID[userID]
	if !userExists {
		s.store.mu.RUnlock()
		return 0, ErrUserNotFound
	}

	item, itemExists := s.store.itemsByID[itemID]
	if !itemExists {
		s.store.mu.RUnlock()
		return 0, ErrItemNotFound
	}

	if item.Status != models.ItemStatusActive {
		s.store.mu.RUnlock()
		return 0, ErrAuctionNotActive
	}

	s.store.mu.RUnlock()

	// Mutations
	s.store.mu.Lock()

	userStack := s.store.getOrCreateUserItemBidStackLocked(userID, itemID)
	lastBidID, ok := userStack.Pop()
	if !ok {
		s.store.mu.Unlock()
		return 0, ErrNoBidToRetract
	}

	lastBid, bidExists := s.store.bidsByID[lastBidID]
	if !bidExists {
		s.store.mu.Unlock()
		return 0, ErrBidNotFound
	}

	if lastBid.UserID != userID {
		s.store.mu.Unlock()
		return 0, ErrBidOwnershipMismatch
	}

	if lastBid.ItemID != itemID {
		s.store.mu.Unlock()
		return 0, ErrBidItemMismatch
	}

	if lastBid.IsRetracted {
		s.store.mu.Unlock()
		return 0, ErrBidAlreadyRetracted
	}

	// Mark bid inactive
	lastBid.IsRetracted = true

	// Optional: keep user.ActiveBids consistent if you still use it
	user := s.store.usersByID[userID]
	user.ActiveBids = removeBidID(user.ActiveBids, lastBidID)

	// Recompute current winner lazily from heap top
	h := s.store.getOrCreateItemHeapLocked(itemID)
	for {
		top, ok := h.Peek()
		if !ok {
			item.CurrentBid = nil
			break
		}
		if top.IsRetracted {
			_, _ = h.Pop()
			continue
		}
		item.CurrentBid = top
		break
	}

	s.store.mu.Unlock()

	s.store.publishToWatchers(itemID, BidEvent{
		ItemID:    itemID,
		BidID:     lastBidID,
		Action:    BidEventRetracted,
		Timestamp: now,
	})

	return lastBidID, nil
}

func (s *AuctionService) EndAuction(itemID models.ItemID) (*models.Bid, error) {
	itemLock := s.store.getItemLock(itemID)
	itemLock.Lock()
	defer itemLock.Unlock()

	now := time.Now().UTC()

	s.store.mu.Lock()

	item, exists := s.store.itemsByID[itemID]
	if !exists {
		s.store.mu.Unlock()
		return nil, ErrItemNotFound
	}
	if item.Status != models.ItemStatusActive {
		s.store.mu.Unlock()
		return nil, ErrAuctionNotActive
	}

	h := s.store.getOrCreateItemHeapLocked(itemID)
	for {
		top, ok := h.Peek()
		if !ok {
			item.CurrentBid = nil
			break
		}
		if top.IsRetracted {
			_, _ = h.Pop()
			continue
		}
		item.CurrentBid = top
		break
	}

	item.Status = models.ItemStatusEnded
	winner := item.CurrentBid

	winnerBidID := models.BidID(0)
	if winner != nil {
		winnerBidID = winner.ID
	}
	s.store.mu.Unlock()

	s.store.publishToWatchers(itemID, BidEvent{
		ItemID:    itemID,
		BidID:     winnerBidID,
		Action:    BidEventEnded,
		Timestamp: now,
	})

	return winner, nil
}

func (s *AuctionService) GetItem(itemID models.ItemID) (*models.Item, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	item, ok := s.store.itemsByID[itemID]
	if !ok {
		return nil, ErrItemNotFound
	}
	c := *item
	if item.CurrentBid != nil {
		b := *item.CurrentBid
		c.CurrentBid = &b
	}
	c.BidHistory = slices.Clone(item.BidHistory)
	return &c, nil
}

func (s *AuctionService) GetItemBids(itemID models.ItemID) ([]*models.Bid, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	item, ok := s.store.itemsByID[itemID]
	if !ok {
		return nil, ErrItemNotFound
	}

	bids := make([]*models.Bid, 0, len(item.BidHistory))
	for _, bidID := range item.BidHistory {
		b, exists := s.store.bidsByID[bidID]
		if !exists {
			continue
		}
		copyBid := *b
		bids = append(bids, &copyBid)
	}
	return bids, nil
}

func (s *AuctionService) GetStats() Stats {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	stats := Stats{
		TotalUsers: len(s.store.usersByID),
		TotalItems: len(s.store.itemsByID),
		TotalBids:  len(s.store.bidsByID),
	}
	var topBidSum float64
	var topBidCount int
	for _, item := range s.store.itemsByID {
		switch item.Status {
		case models.ItemStatusActive:
			stats.ActiveItems++
		case models.ItemStatusEnded:
			stats.EndedItems++
		case models.ItemStatusCancelled:
			stats.CancelledItems++
		}
		if item.CurrentBid != nil && !item.CurrentBid.IsRetracted {
			topBidSum += item.CurrentBid.Amount
			topBidCount++
		}
	}
	for _, bid := range s.store.bidsByID {
		if bid.IsRetracted {
			stats.RetractedBids++
		}
	}
	for _, watchers := range s.store.watchersByItemID {
		stats.ActiveWatchers += len(watchers)
	}
	if topBidCount > 0 {
		stats.AverageTopBid = topBidSum / float64(topBidCount)
	}
	return stats
}

func (s *AuctionService) SubscribeItem(itemID models.ItemID) (chan BidEvent, error) {
	s.store.mu.RLock()
	_, exists := s.store.itemsByID[itemID]
	s.store.mu.RUnlock()
	if !exists {
		return nil, ErrItemNotFound
	}

	watcher := make(chan BidEvent, defaultWatcherChannelSize)
	s.store.addWatcher(itemID, watcher)
	return watcher, nil
}

func (s *AuctionService) UnsubscribeItem(itemID models.ItemID, ch chan BidEvent) {
	s.store.removeWatcher(itemID, ch)
}

func removeBidID(bids []models.BidID, target models.BidID) []models.BidID {
	for i := len(bids) - 1; i >= 0; i-- {
		if bids[i] != target {
			continue
		}
		return append(bids[:i], bids[i+1:]...)
	}
	return bids
}

func normalizeCategoryPath(path []string) []string {
	out := make([]string, 0, len(path))
	for _, category := range path {
		trimmed := strings.TrimSpace(category)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}
