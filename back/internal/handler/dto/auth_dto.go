package dto

import "time"

type AuthResponse struct {
	Token string    `json:"token"`
	ExpAt time.Time `json:"expires_at"`
}
