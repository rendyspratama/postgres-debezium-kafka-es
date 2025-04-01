package models

import (
	"errors"
	"time"
)

type Category struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      int       `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
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
