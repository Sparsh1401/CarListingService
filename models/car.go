package models

import "time"

type Car struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Price     string    `json:"price"`
	Currency  string    `json:"currency"`
	Year      string    `json:"year"`
	Mileage   string    `json:"mileage"`
	Location  string    `json:"location"`
	Link      string    `json:"link"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
