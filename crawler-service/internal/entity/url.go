package entity

// Url represents the domain model for a URL
type Url struct {
	ID          string `json:"id"`
	Url         string `json:"url"`
	Method      string `json:"method"`
	Description string `json:"description"`
	Queue       string `json:"queue"`
	Domain      string `json:"domain"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
