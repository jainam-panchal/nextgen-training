package models

type Rider struct {
	ID            string
	Name          string
	Location      Location
	PaymentMethod string
	RideHistory   []string
}
