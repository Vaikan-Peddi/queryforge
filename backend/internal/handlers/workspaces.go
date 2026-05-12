package handlers

import (
	"net/http"

	appmw "queryforge/backend/internal/middleware"
	"queryforge/backend/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type WorkspaceHandler struct {
	service *services.WorkspaceService
	schema  *services.SchemaService
}

func NewWorkspaceHandler(service *services.WorkspaceService, schema *services.SchemaService) *WorkspaceHandler {
	return &WorkspaceHandler{service: service, schema: schema}
}

type workspaceRequest struct {
	Name string `json:"name"`
}

func (h *WorkspaceHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := appmw.UserID(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	workspaces, err := h.service.List(r.Context(), userID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"workspaces": workspaces})
}

func (h *WorkspaceHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := appmw.UserID(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	var req workspaceRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	workspace, err := h.service.Create(r.Context(), userID, req.Name)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, workspace)
}

func (h *WorkspaceHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, workspaceID, ok := ids(w, r)
	if !ok {
		return
	}
	workspace, err := h.service.GetOwned(r.Context(), userID, workspaceID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, workspace)
}

func (h *WorkspaceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, workspaceID, ok := ids(w, r)
	if !ok {
		return
	}
	if err := h.service.Delete(r.Context(), userID, workspaceID); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *WorkspaceHandler) Upload(w http.ResponseWriter, r *http.Request) {
	userID, workspaceID, ok := ids(w, r)
	if !ok {
		return
	}
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart upload")
		return
	}
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file field is required")
		return
	}
	_ = file.Close()
	workspace, err := h.service.UploadSQLite(r.Context(), userID, workspaceID, fileHeader)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, workspace)
}

func (h *WorkspaceHandler) Schema(w http.ResponseWriter, r *http.Request) {
	userID, workspaceID, ok := ids(w, r)
	if !ok {
		return
	}
	workspace, err := h.service.GetOwned(r.Context(), userID, workspaceID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	if workspace.SQLiteFilePath == "" {
		writeError(w, http.StatusBadRequest, "workspace has no uploaded database")
		return
	}
	schema, err := h.schema.Inspect(r.Context(), workspace.SQLiteFilePath)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, schema)
}

func ids(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
	userID, err := appmw.UserID(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return uuid.Nil, uuid.Nil, false
	}
	workspaceID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid workspace id")
		return uuid.Nil, uuid.Nil, false
	}
	return userID, workspaceID, true
}
