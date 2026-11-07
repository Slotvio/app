package domain

import "time"

type SchedulePeriod struct {
	ID        int64
	MasterID  int64
	DayOfWeek int8
	StartTime time.Time
	EndTime   time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
