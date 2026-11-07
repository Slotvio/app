package domain

import "time"

type Booking struct {
	ID             int
	MasterID       int
	ClientID       int
	ServiceID      int
	StartAt        time.Time
	EndAt          time.Time
	Status         string
	ClientComment  string
	ReminderSentAt time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
