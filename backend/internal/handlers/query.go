package handlers

import (
	"net/http"
	"strconv"

	appmw "queryforge/backend/internal/middleware"
	"queryforge/backend/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type QueryHandler struct {
	service *services.QueryService
}

func NewQueryHandler(service *services.QueryService) *QueryHandler {
	return &QueryHandler{service: service}
}

type queryRequest struct {
	Question string `json:"question"`
	SQL      string `json:"sql"`
}

func (h *QueryHandler) Generate(w http.ResponseWriter, r *http.Request) {
	userID, workspaceID, ok := ids(w, r)
	if !ok {
		return
	}
	var req queryRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	resp, err := h.service.Generate(r.Context(), userID, workspaceID, req.Question)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *QueryHandler) Execute(w http.ResponseWriter, r *http.Request) {
	userID, workspaceID, ok := ids(w, r)
	if !ok {
		return
	}
	var req queryRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	resp, err := h.service.Execute(r.Context(), userID, workspaceID, req.SQL)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *QueryHandler) ListHistory(w http.ResponseWriter, r *http.Request) {
	userID, workspaceID, ok := ids(w, r)
	if !ok {
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	history, err := h.service.ListHistory(r.Context(), userID, workspaceID, limit, offset)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"history": history})
}

func (h *QueryHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	userID, err := appmw.UserID(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	historyID, err := uuid.Parse(chi.URLParam(r, "historyId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid history id")
		return
	}
	history, err := h.service.GetHistory(r.Context(), userID, historyID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, history)
}
