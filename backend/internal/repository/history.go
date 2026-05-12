package repository

import (
	"context"
	"errors"

	"queryforge/backend/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HistoryRepository struct {
	pool *pgxpool.Pool
}

func NewHistoryRepository(pool *pgxpool.Pool) *HistoryRepository {
	return &HistoryRepository{pool: pool}
}

type HistoryCreate struct {
	WorkspaceID  uuid.UUID
	UserID       uuid.UUID
	Question     *string
	GeneratedSQL *string
	ExecutedSQL  *string
	Explanation  *string
	Status       string
	ErrorMessage *string
	ExecutionMS  *int64
}

func (r *HistoryRepository) Create(ctx context.Context, input HistoryCreate) (models.QueryHistory, error) {
	var h models.QueryHistory
	err := r.pool.QueryRow(ctx, `
		INSERT INTO query_history(workspace_id, user_id, question, generated_sql, executed_sql, explanation, status, error_message, execution_ms)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id, workspace_id, user_id, question, generated_sql, executed_sql, explanation, status, error_message, execution_ms, created_at
	`, input.WorkspaceID, input.UserID, input.Question, input.GeneratedSQL, input.ExecutedSQL, input.Explanation, input.Status, input.ErrorMessage, input.ExecutionMS).
		Scan(&h.ID, &h.WorkspaceID, &h.UserID, &h.Question, &h.GeneratedSQL, &h.ExecutedSQL, &h.Explanation, &h.Status, &h.ErrorMessage, &h.ExecutionMS, &h.CreatedAt)
	return h, err
}

func (r *HistoryRepository) List(ctx context.Context, userID, workspaceID uuid.UUID, limit, offset int) ([]models.QueryHistory, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, workspace_id, user_id, question, generated_sql, executed_sql, explanation, status, error_message, execution_ms, created_at
		FROM query_history
		WHERE user_id=$1 AND workspace_id=$2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`, userID, workspaceID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.QueryHistory
	for rows.Next() {
		var h models.QueryHistory
		if err := rows.Scan(&h.ID, &h.WorkspaceID, &h.UserID, &h.Question, &h.GeneratedSQL, &h.ExecutedSQL, &h.Explanation, &h.Status, &h.ErrorMessage, &h.ExecutionMS, &h.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

func (r *HistoryRepository) FindOwned(ctx context.Context, userID, historyID uuid.UUID) (models.QueryHistory, error) {
	var h models.QueryHistory
	err := r.pool.QueryRow(ctx, `
		SELECT id, workspace_id, user_id, question, generated_sql, executed_sql, explanation, status, error_message, execution_ms, created_at
		FROM query_history
		WHERE user_id=$1 AND id=$2
	`, userID, historyID).Scan(&h.ID, &h.WorkspaceID, &h.UserID, &h.Question, &h.GeneratedSQL, &h.ExecutedSQL, &h.Explanation, &h.Status, &h.ErrorMessage, &h.ExecutionMS, &h.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return h, ErrNotFound
	}
	return h, err
}
