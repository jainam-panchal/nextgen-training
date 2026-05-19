// Package models
package models

import "time"

type Product struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Category  string    `json:"category"`
	Price     float64   `json:"price"`
	Rating    float64   `json:"rating"`
	Stock     int       `json:"stock"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
}
