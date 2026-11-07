package domain

import "time"

type Service struct {
	ID              int
	MasterID        int
	Name            string
	Description     string
	DurationMinutes int
	Price           int
	IsActive        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
