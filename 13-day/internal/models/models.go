package models

import "time"

type UserID int
type ItemID int
type BidID int

type ItemStatus string

const (
	ItemStatusActive    ItemStatus = "active"
	ItemStatusEnded     ItemStatus = "ended"
	ItemStatusCancelled ItemStatus = "cancelled"
)

type Bid struct {
	ID          BidID
	ItemID      ItemID
	UserID      UserID
	Amount      float64
	IsRetracted bool
	Timestamp   time.Time
}

type User struct {
	ID         UserID
	Name       string
	Balance    float64
	ActiveBids []BidID
}

type Item struct {
	ID          ItemID
	Name        string
	Category    string
	Description string
	SellerID    UserID
	StartPrice  float64
	CurrentBid  *Bid
	BidHistory  []BidID
	StartTime   time.Time
	EndTime     time.Time
	Status      ItemStatus
}
