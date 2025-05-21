package domain

// Url represents the domain model for a URL
type Url struct {
	ID          string
	Url         string
	Description string
	Queue       string
	Domain      string
	IsActive    bool
	CreatedAt   string
	UpdatedAt   string
}
