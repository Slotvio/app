package domain

import "time"

type SchedulePeriod struct {
	ID        int
	MasterID  int
	DayOfWeek int8
	StartTime time.Time
	EndTime   time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
