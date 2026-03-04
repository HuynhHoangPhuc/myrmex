package entity

import "time"

// Notification represents a stored notification record.
type Notification struct {
	ID        string
	UserID    string
	Type      string         // e.g. "schedule.published", "enrollment.approved"
	Channel   string         // "in_app" | "email"
	Title     string
	Body      string
	Metadata  map[string]any // resource_type, resource_id, link, etc.
	ReadAt    *time.Time
	CreatedAt time.Time
}
