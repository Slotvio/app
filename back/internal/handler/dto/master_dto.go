package dto

type CreateProfileRequest struct {
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Category    string `json:"category"`
	Description string `json:"description,omitempty"`
	Timezone    string `json:"timezone"`
}
