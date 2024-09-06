package models

import "time"

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username" validate:"required,min=3,max=30"`
	Password  string    `json:"password,omitempty" validate:"required,min=8"`
	Email     string    `json:"email" validate:"required,email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
