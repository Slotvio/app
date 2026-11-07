package domain

import "time"

type User struct {
	ID         int
	TelegramID int
	FirstName  string
	LastName   string
	Username   string
	Phone      string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
