package handlers

import (
	"net/http"
	"time"

	"github.com/rendyspratama/digital-discovery/api/utils"
)

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	utils.WriteSuccess(w, response)
}
