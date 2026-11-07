package domain

import "time"

type Service struct {
	ID              int64
	MasterID        int64
	Name            string
	Description     string
	DurationMinutes int
	Price           int
	IsActive        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
