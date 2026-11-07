package domain

import "time"

type Master struct {
	ID                int
	UserID            int
	Slug              string
	Title             string
	Description       string
	Timezone          string
	SubscriptionUntil time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
