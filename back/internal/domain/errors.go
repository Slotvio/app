package domain

import "errors"

var (
	ErrNotFound        = errors.New("entity not found")
	ErrBookingConflict = errors.New("booking overlaps an active booking")
)
