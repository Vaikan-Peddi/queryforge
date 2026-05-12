package repository

import (
	"context"
	"errors"

	"queryforge/backend/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WorkspaceRepository struct {
	pool *pgxpool.Pool
}

func NewWorkspaceRepository(pool *pgxpool.Pool) *WorkspaceRepository {
	return &WorkspaceRepository{pool: pool}
}

func (r *WorkspaceRepository) List(ctx context.Context, userID uuid.UUID) ([]models.Workspace, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, name, db_type, COALESCE(sqlite_file_path, ''), sqlite_file_path IS NOT NULL, created_at, updated_at
		FROM workspaces
		WHERE user_id=$1
		ORDER BY updated_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Workspace
	for rows.Next() {
		var w models.Workspace
		if err := rows.Scan(&w.ID, &w.UserID, &w.Name, &w.DBType, &w.SQLiteFilePath, &w.HasDatabase, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

func (r *WorkspaceRepository) Create(ctx context.Context, userID uuid.UUID, name, dbType string) (models.Workspace, error) {
	var w models.Workspace
	err := r.pool.QueryRow(ctx, `
		INSERT INTO workspaces(user_id, name, db_type)
		VALUES($1, $2, $3)
		RETURNING id, user_id, name, db_type, COALESCE(sqlite_file_path, ''), sqlite_file_path IS NOT NULL, created_at, updated_at
	`, userID, name, dbType).Scan(&w.ID, &w.UserID, &w.Name, &w.DBType, &w.SQLiteFilePath, &w.HasDatabase, &w.CreatedAt, &w.UpdatedAt)
	return w, err
}

func (r *WorkspaceRepository) FindOwned(ctx context.Context, userID, workspaceID uuid.UUID) (models.Workspace, error) {
	var w models.Workspace
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, name, db_type, COALESCE(sqlite_file_path, ''), sqlite_file_path IS NOT NULL, created_at, updated_at
		FROM workspaces
		WHERE user_id=$1 AND id=$2
	`, userID, workspaceID).Scan(&w.ID, &w.UserID, &w.Name, &w.DBType, &w.SQLiteFilePath, &w.HasDatabase, &w.CreatedAt, &w.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return w, ErrNotFound
	}
	return w, err
}

func (r *WorkspaceRepository) UpdateSQLitePath(ctx context.Context, userID, workspaceID uuid.UUID, path string) (models.Workspace, error) {
	var w models.Workspace
	err := r.pool.QueryRow(ctx, `
		UPDATE workspaces
		SET sqlite_file_path=$3, updated_at=now()
		WHERE user_id=$1 AND id=$2
		RETURNING id, user_id, name, db_type, COALESCE(sqlite_file_path, ''), sqlite_file_path IS NOT NULL, created_at, updated_at
	`, userID, workspaceID, path).Scan(&w.ID, &w.UserID, &w.Name, &w.DBType, &w.SQLiteFilePath, &w.HasDatabase, &w.CreatedAt, &w.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return w, ErrNotFound
	}
	return w, err
}

func (r *WorkspaceRepository) Delete(ctx context.Context, userID, workspaceID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM workspaces WHERE user_id=$1 AND id=$2`, userID, workspaceID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
