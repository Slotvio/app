package dto

import "time"

type AvailableSlotsResponse struct {
	Date     string          `json:"date"`
	Timezone string          `json:"timezone"`
	Slots    []AvailableSlot `json:"slots"`
}

type AvailableSlot struct {
	StartAt   time.Time `json:"start_at"`
	LocalTime string    `json:"local_time"`
}

type CreateBookingRequest struct {
	MasterSlug    string    `json:"master_slug"`
	ServiceID     int64     `json:"service_id"`
	StartAt       time.Time `json:"start_at"`
	ClientComment string    `json:"client_comment,omitempty"`
}

type CreateBookingResponse struct {
	ID              int64     `json:"id"`
	MasterID        int64     `json:"master_id"`
	ServiceID       int64     `json:"service_id"`
	ServiceName     string    `json:"service_name"`
	DurationMinutes int       `json:"duration_minutes"`
	Price           int       `json:"price"`
	StartAt         time.Time `json:"start_at"`
	EndAt           time.Time `json:"end_at"`
	Status          string    `json:"status"`
}
