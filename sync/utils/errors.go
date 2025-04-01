package utils

import "fmt"

type SyncError struct {
	Code       string
	Message    string
	Err        error
	StatusCode int    // HTTP status code equivalent
	Operation  string // The operation being performed
	Entity     string // The entity being processed
}

func (e *SyncError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v (operation: %s, entity: %s)",
			e.Code, e.Message, e.Err, e.Operation, e.Entity)
	}
	return fmt.Sprintf("[%s] %s (operation: %s, entity: %s)",
		e.Code, e.Message, e.Operation, e.Entity)
}

// Error codes with categories
const (
	// Kafka related errors
	ErrCodeKafkaConnection  = "SYNC_KAFKA_001"
	ErrCodeKafkaConsumer    = "SYNC_KAFKA_002"
	ErrCodeKafkaDeserialize = "SYNC_KAFKA_003"
	ErrCodeKafkaCommit      = "SYNC_KAFKA_004"

	// Elasticsearch related errors
	ErrCodeESConnection = "SYNC_ES_001"
	ErrCodeESIndex      = "SYNC_ES_002"
	ErrCodeESTemplate   = "SYNC_ES_003"
	ErrCodeESLifecycle  = "SYNC_ES_004"
	ErrCodeESQuery      = "SYNC_ES_005"
	ErrCodeESConflict   = "SYNC_ES_006"
	ErrCodeESTimeout    = "SYNC_ES_007"

	// Data related errors
	ErrCodeInvalidPayload = "SYNC_DATA_001"
	ErrCodeDataValidation = "SYNC_DATA_002"
	ErrCodeDataTransform  = "SYNC_DATA_003"
	ErrCodeDataConflict   = "SYNC_DATA_004"

	// Retry related errors
	ErrCodeRetryExhausted = "SYNC_RETRY_001"
	ErrCodeRetryTimeout   = "SYNC_RETRY_002"
	ErrCodeRetryCircuit   = "SYNC_RETRY_003"

	// System errors
	ErrCodeSystemConfig   = "SYNC_SYS_001"
	ErrCodeSystemResource = "SYNC_SYS_002"
	ErrCodeSystemState    = "SYNC_SYS_003"

	// Operation specific errors
	ErrCodeCreateFailed    = "SYNC_OP_001"
	ErrCodeUpdateFailed    = "SYNC_OP_002"
	ErrCodeDeleteFailed    = "SYNC_OP_003"
	ErrCodeVersionConflict = "SYNC_OP_004"

	// Validation errors
	ErrCodeValidationFailed = "SYNC_VAL_001"
	ErrCodeSchemaInvalid    = "SYNC_VAL_002"

	// Connection errors
	ErrCodeConnectionFailed = "SYNC_CONN_001"
	ErrCodeTimeout          = "SYNC_CONN_002"

	// Kafka specific errors
	ErrCodeKafkaConsumerInit = "SYNC_KAFKA_005"
	ErrCodeKafkaGroupJoin    = "SYNC_KAFKA_006"
	ErrCodeKafkaRebalance    = "SYNC_KAFKA_007"
)

// Error constructors with enhanced context
func NewSyncError(code string, msg string, err error, operation string, entity string) *SyncError {
	return &SyncError{
		Code:      code,
		Message:   msg,
		Err:       err,
		Operation: operation,
		Entity:    entity,
	}
}

// Specific error constructors
func NewKafkaError(msg string, err error) *SyncError {
	return &SyncError{
		Code:       ErrCodeKafkaDeserialize,
		Message:    msg,
		Err:        err,
		StatusCode: 500,
		Operation:  "kafka",
		Entity:     "message",
	}
}

func NewESError(code string, msg string, err error, operation string, index string) *SyncError {
	return &SyncError{
		Code:       code,
		Message:    msg,
		Err:        err,
		StatusCode: 500,
		Operation:  operation,
		Entity:     fmt.Sprintf("elasticsearch:%s", index),
	}
}

func NewDataError(code string, msg string, err error, dataType string) *SyncError {
	return &SyncError{
		Code:       code,
		Message:    msg,
		Err:        err,
		StatusCode: 400,
		Operation:  "data_validation",
		Entity:     dataType,
	}
}

// ... other error constructors

func NewESIndexError(msg string, err error) *SyncError {
	return &SyncError{
		Code:       ErrCodeESIndex,
		Message:    msg,
		Err:        err,
		StatusCode: 500,
		Operation:  "index",
		Entity:     "elasticsearch",
	}
}

// Add Kafka-specific error constructor
func NewKafkaConsumerError(msg string, err error, operation string) *SyncError {
	return &SyncError{
		Code:       ErrCodeKafkaConsumer,
		Message:    msg,
		Err:        err,
		StatusCode: 500,
		Operation:  operation,
		Entity:     "kafka_consumer",
	}
}

// Add IsRetryableError function to determine if an error should be retried
func IsRetryableError(err error) bool {
	if syncErr, ok := err.(*SyncError); ok {
		switch syncErr.Code {
		case ErrCodeESIndex, ErrCodeESConnection, ErrCodeKafkaDeserialize:
			return true
		default:
			return false
		}
	}
	return false
}
