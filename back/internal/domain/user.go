package domain

import "time"

type User struct {
	ID         int64
	TelegramID int64
	FirstName  string
	LastName   string
	Username   string
	Phone      string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
