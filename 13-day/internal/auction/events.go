package auction

import (
	"realtime-auction/internal/models"
	"time"
)

type BidEventAction string

const (
	BidEventPlaced    BidEventAction = "placed"
	BidEventRetracted BidEventAction = "retracted"
)

type BidEvent struct {
	ItemID    models.ItemID
	BidID     models.BidID
	Action    BidEventAction
	Timestamp time.Time
}
