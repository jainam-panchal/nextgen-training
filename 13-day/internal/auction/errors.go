package auction

import "errors"

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrItemNotFound        = errors.New("item not found")
	ErrBidNotFound         = errors.New("bid not found")
	ErrCategoryNotFound    = errors.New("category not found")
	ErrInvalidBidAmount    = errors.New("invalid bid amount")
	ErrAuctionNotActive    = errors.New("auction is not active")
	ErrSellerCannotBid     = errors.New("seller cannot place bids on their own items")
	ErrInsufficientBalance = errors.New("user has insufficient balance to place bid")

	ErrBidOwnershipMismatch = errors.New("bid does not belong to the specified user")
	ErrBidItemMismatch      = errors.New("bid does not belong to the specified item")
	ErrBidAlreadyRetracted  = errors.New("bid has already been retracted")
	ErrNoBidToRetract       = errors.New("no bid to retract")
)
