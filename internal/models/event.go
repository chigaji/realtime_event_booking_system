package models

import "time"

type Event struct {
	ID           int    `json:"id"`
	Name         string `json:"name" validate:"required,min=3,max=100"`
	Description  string `json:"description" validate:"required,min=3"`
	TotalTickets int    `json:"total_tickets" validate:"required,min=1"`
	// AvailableTickets int       `json:"available_tickets" validate:"required,min=1"`
	BookedTickets int       `json:"booked_tickets"`
	EventDate     time.Time `json:"event_date"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
