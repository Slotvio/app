package domain

import "time"

type SchedulePeriod struct {
	ID        int
	MasterID  int
	DayOfWeek int8
	StartAt   time.Time
	EndAt     time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
