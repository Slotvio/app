package domain

import "time"

type Booking struct {
	ID             int64
	MasterID       int64
	ClientID       int64
	ServiceID      int64
	StartAt        time.Time
	EndAt          time.Time
	Status         string
	ClientComment  string
	ReminderSentAt time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
