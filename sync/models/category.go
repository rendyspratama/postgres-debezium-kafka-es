package models

import (
	"errors"
	"time"
)

type Category struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Status      int64      `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Version     int64      `json:"version"`
	SyncStatus  SyncStatus `json:"sync_status"`
	LastSync    time.Time  `json:"last_sync"`
}

type CategoryOperation struct {
	Operation string    `json:"operation"`
	Payload   Category  `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
}

// Validate checks if the category data is valid
func (c *Category) Validate() error {
	if c.Name == "" {
		return errors.New("name is required")
	}
	// Make description optional by removing its validation
	if c.Status < 0 {
		return errors.New("status must be non-negative")
	}
	return nil
}
