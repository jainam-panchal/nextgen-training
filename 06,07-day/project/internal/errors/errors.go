package errors

import stderrors "errors"

var (
	ErrDriverNotFound      = stderrors.New("driver not found")
	ErrDuplicateDriver     = stderrors.New("driver already exists")
	ErrNoDriverFound       = stderrors.New("no available driver found")
	ErrInvalidDriver       = stderrors.New("invalid driver")
	ErrInvalidLocation     = stderrors.New("invalid location")
	ErrInvalidStatus       = stderrors.New("invalid status")
	ErrInvalidStatusChange = stderrors.New("invalid driver status change")
	ErrInvalidRating       = stderrors.New("invalid rating")

	ErrRiderNotFound  = stderrors.New("rider not found")
	ErrDuplicateRider = stderrors.New("rider already exists")
	ErrInvalidRider   = stderrors.New("invalid rider")
	ErrInvalidRide    = stderrors.New("invalid ride")

	ErrEmptyQueue         = stderrors.New("ride request queue is empty")
	ErrInvalidRideRequest = stderrors.New("invalid ride request")

	ErrEmptyLinkedList       = stderrors.New("linked list is empty")
	ErrInvalidLinkedListNode = stderrors.New("invalid linked list node")

	ErrRideNotFound  = stderrors.New("ride not found")
	ErrDuplicateRide = stderrors.New("ride already exists")

	ErrRideRequestExpired = stderrors.New("ride request expired")
	ErrNoDriverAvailable  = stderrors.New("no available driver")
)
