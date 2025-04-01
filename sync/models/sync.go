package models

import "time"

type SyncStatus string

const (
	SyncStatusPending  SyncStatus = "PENDING"
	SyncStatusSuccess  SyncStatus = "SUCCESS"
	SyncStatusFailed   SyncStatus = "FAILED"
	SyncStatusRetrying SyncStatus = "RETRYING"
)

// Add operation constants
const (
	OperationCreate = "CREATE"
	OperationUpdate = "UPDATE"
	OperationDelete = "DELETE"
)

type SyncRecord struct {
	ID           string     `json:"id"`
	EntityType   string     `json:"entity_type"`
	EntityID     string     `json:"entity_id"`
	Operation    string     `json:"operation"`
	Status       SyncStatus `json:"status"`
	ErrorMessage string     `json:"error_message,omitempty"`
	RetryCount   int        `json:"retry_count"`
	LastRetry    *time.Time `json:"last_retry,omitempty"`
	NextRetry    *time.Time `json:"next_retry,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// Add validation method
func IsValidOperation(op string) bool {
	switch op {
	case OperationCreate, OperationUpdate, OperationDelete:
		return true
	default:
		return false
	}
}

func (s *SyncRecord) MarkAsFailed(err error, retryDelay time.Duration) {
	now := time.Now()
	s.Status = SyncStatusFailed
	s.ErrorMessage = err.Error()
	s.RetryCount++
	s.LastRetry = &now
	nextRetry := now.Add(retryDelay)
	s.NextRetry = &nextRetry
	s.UpdatedAt = now
}

func (s *SyncRecord) MarkAsSuccess() {
	now := time.Now()
	s.Status = SyncStatusSuccess
	s.ErrorMessage = ""
	s.UpdatedAt = now
	s.LastRetry = nil
	s.NextRetry = nil
}
