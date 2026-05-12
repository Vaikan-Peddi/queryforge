package handlers

import (
	"net/http"

	"queryforge/backend/internal/services"
)

type AIHandler struct {
	client *services.AIClient
}

func NewAIHandler(client *services.AIClient) *AIHandler {
	return &AIHandler{client: client}
}

func (h *AIHandler) Health(w http.ResponseWriter, r *http.Request) {
	health, err := h.client.Health(r.Context())
	if err != nil {
		writeError(w, http.StatusBadGateway, "ai service unavailable")
		return
	}
	writeJSON(w, http.StatusOK, health)
}
