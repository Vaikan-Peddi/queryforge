package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"queryforge/backend/internal/config"
	"queryforge/backend/internal/models"
	"queryforge/backend/internal/repository"

	"github.com/google/uuid"
)

type WorkspaceService struct {
	cfg    config.Config
	repo   *repository.WorkspaceRepository
	schema *SchemaService
}

func NewWorkspaceService(cfg config.Config, repo *repository.WorkspaceRepository, schema *SchemaService) *WorkspaceService {
	return &WorkspaceService{cfg: cfg, repo: repo, schema: schema}
}

func (s *WorkspaceService) List(ctx context.Context, userID uuid.UUID) ([]models.Workspace, error) {
	return s.repo.List(ctx, userID)
}

func (s *WorkspaceService) Create(ctx context.Context, userID uuid.UUID, name string) (models.Workspace, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return models.Workspace{}, errors.New("workspace name is required")
	}
	return s.repo.Create(ctx, userID, name, "sqlite")
}

func (s *WorkspaceService) GetOwned(ctx context.Context, userID, workspaceID uuid.UUID) (models.Workspace, error) {
	return s.repo.FindOwned(ctx, userID, workspaceID)
}

func (s *WorkspaceService) Delete(ctx context.Context, userID, workspaceID uuid.UUID) error {
	return s.repo.Delete(ctx, userID, workspaceID)
}

func (s *WorkspaceService) UploadSQLite(ctx context.Context, userID, workspaceID uuid.UUID, header *multipart.FileHeader) (models.Workspace, error) {
	if header == nil {
		return models.Workspace{}, errors.New("file is required")
	}
	if header.Size <= 0 || header.Size > s.cfg.MaxUploadBytes {
		return models.Workspace{}, fmt.Errorf("file must be between 1 byte and %d bytes", s.cfg.MaxUploadBytes)
	}
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".sqlite" && ext != ".sqlite3" && ext != ".db" {
		return models.Workspace{}, errors.New("only .sqlite, .sqlite3, and .db files are supported")
	}
	if _, err := s.repo.FindOwned(ctx, userID, workspaceID); err != nil {
		return models.Workspace{}, err
	}

	src, err := header.Open()
	if err != nil {
		return models.Workspace{}, err
	}
	defer src.Close()

	dir := filepath.Join(s.cfg.StorageDir, userID.String(), workspaceID.String())
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return models.Workspace{}, err
	}
	targetPath := filepath.Join(dir, "database.sqlite")
	tmpPath := targetPath + ".tmp"
	dst, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o640)
	if err != nil {
		return models.Workspace{}, err
	}
	if _, err := io.Copy(dst, io.LimitReader(src, s.cfg.MaxUploadBytes+1)); err != nil {
		_ = dst.Close()
		return models.Workspace{}, err
	}
	if err := dst.Close(); err != nil {
		return models.Workspace{}, err
	}
	if err := s.schema.ValidateSQLiteFile(ctx, tmpPath); err != nil {
		_ = os.Remove(tmpPath)
		return models.Workspace{}, fmt.Errorf("invalid SQLite database: %w", err)
	}
	if err := os.Rename(tmpPath, targetPath); err != nil {
		return models.Workspace{}, err
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return s.repo.UpdateSQLitePath(ctx, userID, workspaceID, targetPath)
}
