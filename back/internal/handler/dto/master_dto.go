package dto

import "time"

// HandleCreateProfile
type CreateProfileRequest struct {
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Category    string `json:"category"`
	Description string `json:"description,omitempty"`
	Timezone    string `json:"timezone"`
}

type CreateProfileResponse struct {
	ID                int64     `json:"id"`
	Slug              string    `json:"slug"`
	Title             string    `json:"title"`
	Category          string    `json:"category"`
	Description       string    `json:"description"`
	Timezone          string    `json:"timezone"`
	SubscriptionUntil time.Time `json:"subscription_until"`
}

// HandleOwnProfile
type OwnProfileResponse struct {
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Timezone    string `json:"timezone"`
}

// HandlePublicProfile

type PublicProfileResponse struct {
	Slug        string        `json:"slug"`
	Title       string        `json:"title"`
	Category    string        `json:"category"`
	Description string        `json:"description"`
	Timezone    string        `json:"timezone"`
	Services    []ServiceInfo `json:"services"`
}

// HandleUpdateProfile
type UpdateProfileRequest struct {
	Title       *string `json:"title"`
	Category    *string `json:"category"`
	Description *string `json:"description"`
	Timezone    *string `json:"timezone"`
}

// HandleListServices
type ListServiceResponse struct {
	Services []ServiceInfo `json:"services"`
}

// HandleListServices
type ServiceInfo struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	DurationMinutes int    `json:"duration_minutes"`
	Price           int    `json:"price"`
}

// HandleCreateService

type CreateServiceRequest struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	DurationMinutes int    `json:"duration_minutes"`
	Price           int    `json:"price"`
}

type CreateServiceResponse struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	DurationMinutes int    `json:"duration_minutes"`
	Price           int    `json:"price"`
	IsActive        bool   `json:"is_active"`
}
