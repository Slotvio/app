package dto

// HandleCreateProfile
type CreateProfileRequest struct {
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Category    string `json:"category"`
	Description string `json:"description,omitempty"`
	Timezone    string `json:"timezone"`
}

// HandleOwnProfile and HandlePublicProfile
type ProfileResponse struct {
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Timezone    string `json:"timezone"`
}
