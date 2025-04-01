package utils

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, status int, message string) {
	response := Response{
		Status: "error",
		Error:  message,
	}
	WriteJSON(w, status, response)
}

func WriteSuccess(w http.ResponseWriter, data interface{}) {
	response := Response{
		Status: "success",
		Data:   data,
	}
	WriteJSON(w, http.StatusOK, response)
}

func WriteSuccessWithRequestID(w http.ResponseWriter, data interface{}, requestID string) {
	response := map[string]interface{}{
		"status":     "success",
		"data":       data,
		"request_id": requestID,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func WriteErrorWithRequestID(w http.ResponseWriter, status int, message string, requestID string) {
	response := map[string]interface{}{
		"status":     "error",
		"message":    message,
		"request_id": requestID,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}
