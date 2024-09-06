package models

import "time"

type Booking struct {
	ID      int `json:"id"`
	UserID  int `json:"user_id" validate:"required"`
	EventID int `json:"event_id" validate:"required"`
	//added
	Quantity  int       `json:"quantity" validate:"required"`
	CreatedAt time.Time `json:"created_at"`
}
