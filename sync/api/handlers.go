package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rendyspratama/digital-discovery/sync/config"
	"github.com/rendyspratama/digital-discovery/sync/services"
	"github.com/rendyspratama/digital-discovery/sync/utils/logger"
)

type Handler struct {
	cfg         *config.Config
	syncService *services.SyncService
	logger      logger.Logger
}

func NewHandler(cfg *config.Config, syncService *services.SyncService, logger logger.Logger) *Handler {
	return &Handler{
		cfg:         cfg,
		syncService: syncService,
		logger:      logger,
	}
}

func (h *Handler) GetSyncMode(w http.ResponseWriter, r *http.Request) {
	status := struct {
		Mode           string `json:"mode"`
		Enabled        bool   `json:"enabled"`
		Status         string `json:"status"`
		CurrentIndex   string `json:"current_index"`
		ConsumerStatus string `json:"consumer_status"`
		ESStatus       string `json:"es_status"`
	}{
		Mode: h.cfg.Sync.Mode,
		Enabled: h.cfg.Sync.Mode == "custom" && h.cfg.Sync.Custom.Enabled ||
			h.cfg.Sync.Mode == "kafka-connect" && h.cfg.Sync.KafkaConnect.Enabled,
		CurrentIndex: h.syncService.GetCurrentIndexName("categories"),
	}

	// Check Elasticsearch health
	if err := h.syncService.HealthCheck(); err != nil {
		status.ESStatus = "unhealthy"
		status.Status = "degraded"
	} else {
		status.ESStatus = "healthy"
		status.Status = "operational"
	}

	// Get consumer status for custom mode
	if h.cfg.Sync.Mode == "custom" {
		if err := h.syncService.HealthCheck(); err != nil {
			status.ConsumerStatus = "unhealthy"
			status.Status = "degraded"
		} else {
			status.ConsumerStatus = "healthy"
		}
	} else if h.cfg.Sync.Mode == "kafka-connect" {
		status.ConsumerStatus = "using-kafka-connect"
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		h.logger.WithError(r.Context(), err, "Failed to encode response", nil)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) UpdateSyncMode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Mode string `json:"mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(r.Context(), err, "Invalid request body", nil)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Mode != "custom" && req.Mode != "kafka-connect" {
		msg := "Invalid mode: must be 'custom' or 'kafka-connect'"
		h.logger.Error(r.Context(), msg, map[string]interface{}{"requested_mode": req.Mode})
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	// Check if requested mode is enabled
	if req.Mode == "custom" && !h.cfg.Sync.Custom.Enabled {
		msg := "Custom sync mode is not enabled"
		h.logger.Error(r.Context(), msg, nil)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	if req.Mode == "kafka-connect" && !h.cfg.Sync.KafkaConnect.Enabled {
		msg := "Kafka Connect mode is not enabled"
		h.logger.Error(r.Context(), msg, nil)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	// Log mode change
	h.logger.Info(r.Context(), "Sync mode change requested", map[string]interface{}{
		"from_mode": h.cfg.Sync.Mode,
		"to_mode":   req.Mode,
	})

	// Update mode in config
	h.cfg.Sync.Mode = req.Mode

	response := map[string]interface{}{
		"message": fmt.Sprintf("Switching to %s mode", req.Mode),
		"status":  "success",
		"mode":    req.Mode,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.WithError(r.Context(), err, "Failed to encode response", nil)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
