package auction

import (
	"realtime-auction/internal/models"
	"strings"
	"time"
)

type AuctionService struct {
	store *AuctionStore
}

func NewAuctionService(store *AuctionStore) *AuctionService {
	return &AuctionService{
		store: store,
	}
}

func (s *AuctionService) CreateUser(name string, balance float64) (models.UserID, error) {
	if name == "" || balance < 0 {
		return 0, ErrInvalidBidAmount
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
	if name == "" || startPrice <= 0 || !endTime.After(startTime) {
		return 0, ErrInvalidBidAmount
	}
	if len(categoryPath) == 0 {
		return 0, ErrCategoryNotFound
	}

	s.store.mu.Lock()
	defer s.store.mu.Unlock()

	if _, ok := s.store.usersByID[sellerID]; !ok {
		return 0, ErrUserNotFound
	}

	if err := s.store.ensureCategoryPathLocked(categoryPath); err != nil {
		return 0, err
	}

	itemID := s.store.generateItemID()
	leafCategory := categoryPath[len(categoryPath)-1]
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

	if amount <= 0 {
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

	history := s.store.getOrCreateItemBidHistoryLocked(itemID)
	history.AddRear(bidID)

	st := s.store.getOrCreateUserItemBidStackLocked(userID, itemID)
	st.Push(bidID)
	user.ActiveBids = append(user.ActiveBids, bidID)

	item.CurrentBid = bid

	ch := s.store.getOrCreateLiveQueueLocked(itemID)

	s.store.mu.Unlock()

	// Non-blocking publish
	select {
	case ch <- BidEvent{
		ItemID:    itemID,
		BidID:     bidID,
		Action:    BidEventPlaced,
		Timestamp: now,
	}:
	default:
	}

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

	// Append retract event in item history list (BidID list design)
	history := s.store.getOrCreateItemBidHistoryLocked(itemID)
	history.AddRear(lastBidID)

	ch := s.store.getOrCreateLiveQueueLocked(itemID)

	s.store.mu.Unlock()

	// Non-blocking publish
	select {
	case ch <- BidEvent{
		ItemID:    itemID,
		BidID:     lastBidID,
		Action:    BidEventRetracted,
		Timestamp: now,
	}:
	default:
		// drop when queue is full
	}

	return lastBidID, nil
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
