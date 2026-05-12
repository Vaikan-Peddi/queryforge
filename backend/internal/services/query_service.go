package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"queryforge/backend/internal/models"
	"queryforge/backend/internal/repository"
	"queryforge/backend/internal/sqlsafe"

	"github.com/google/uuid"
)

type QueryService struct {
	workspaces *repository.WorkspaceRepository
	history    *repository.HistoryRepository
	schema     *SchemaService
	ai         *AIClient
}

func NewQueryService(workspaces *repository.WorkspaceRepository, history *repository.HistoryRepository, schema *SchemaService, ai *AIClient) *QueryService {
	return &QueryService{workspaces: workspaces, history: history, schema: schema, ai: ai}
}

func (s *QueryService) Generate(ctx context.Context, userID, workspaceID uuid.UUID, question string) (GenerateResponse, error) {
	workspace, err := s.workspaces.FindOwned(ctx, userID, workspaceID)
	if err != nil {
		return GenerateResponse{}, err
	}
	if workspace.SQLiteFilePath == "" {
		return GenerateResponse{}, errors.New("workspace has no uploaded database")
	}
	question = strings.TrimSpace(question)
	if question == "" {
		return GenerateResponse{}, errors.New("question is required")
	}
	schema, err := s.schema.Inspect(ctx, workspace.SQLiteFilePath)
	if err != nil {
		return GenerateResponse{}, err
	}
	resp, err := s.ai.GenerateSQL(ctx, GenerateRequest{
		Question: question,
		Schema:   schema,
		SafetyRules: []string{
			"Only SELECT or read-only WITH queries.",
			"No mutation, DDL, PRAGMA, comments, or multiple statements.",
			"Prefer LIMIT 100 unless user asks for a smaller limit.",
		},
	})
	status := "generated"
	var errMsg *string
	if err != nil {
		status = "failed"
		msg := err.Error()
		errMsg = &msg
	}
	if err == nil {
		safeSQL, validationErr := sqlsafe.ValidateAndRewrite(resp.SQL)
		if validationErr != nil {
			status = "failed"
			msg := validationErr.Error()
			errMsg = &msg
			err = validationErr
		} else {
			resp.SQL = safeSQL
		}
	}
	_, _ = s.history.Create(ctx, repository.HistoryCreate{
		WorkspaceID:  workspaceID,
		UserID:       userID,
		Question:     &question,
		GeneratedSQL: stringPtr(resp.SQL),
		Explanation:  stringPtr(resp.Explanation),
		Status:       status,
		ErrorMessage: errMsg,
	})
	return resp, err
}

func (s *QueryService) Execute(ctx context.Context, userID, workspaceID uuid.UUID, rawSQL string) (models.QueryResult, error) {
	workspace, err := s.workspaces.FindOwned(ctx, userID, workspaceID)
	if err != nil {
		return models.QueryResult{}, err
	}
	if workspace.SQLiteFilePath == "" {
		return models.QueryResult{}, errors.New("workspace has no uploaded database")
	}
	safeSQL, err := sqlsafe.ValidateAndRewrite(rawSQL)
	if err != nil {
		msg := err.Error()
		_, _ = s.history.Create(ctx, repository.HistoryCreate{
			WorkspaceID:  workspaceID,
			UserID:       userID,
			ExecutedSQL:  &rawSQL,
			Status:       "failed",
			ErrorMessage: &msg,
		})
		return models.QueryResult{}, err
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	start := time.Now()
	db, err := openReadOnlySQLite(workspace.SQLiteFilePath)
	if err != nil {
		return models.QueryResult{}, err
	}
	defer db.Close()
	if _, err := db.ExecContext(ctx, `PRAGMA query_only=ON`); err != nil {
		return models.QueryResult{}, err
	}
	result, err := executeRows(ctx, db, safeSQL)
	elapsed := time.Since(start).Milliseconds()
	result.ExecutionMS = elapsed
	status := "executed"
	var errMsg *string
	if err != nil {
		status = "failed"
		msg := err.Error()
		errMsg = &msg
	}
	_, _ = s.history.Create(ctx, repository.HistoryCreate{
		WorkspaceID:  workspaceID,
		UserID:       userID,
		ExecutedSQL:  &safeSQL,
		Status:       status,
		ErrorMessage: errMsg,
		ExecutionMS:  &elapsed,
	})
	return result, err
}

func (s *QueryService) ListHistory(ctx context.Context, userID, workspaceID uuid.UUID, limit, offset int) ([]models.QueryHistory, error) {
	if limit <= 0 || limit > 100 {
		limit = 25
	}
	if offset < 0 {
		offset = 0
	}
	if _, err := s.workspaces.FindOwned(ctx, userID, workspaceID); err != nil {
		return nil, err
	}
	return s.history.List(ctx, userID, workspaceID, limit, offset)
}

func (s *QueryService) GetHistory(ctx context.Context, userID, historyID uuid.UUID) (models.QueryHistory, error) {
	return s.history.FindOwned(ctx, userID, historyID)
}

func executeRows(ctx context.Context, db *sql.DB, statement string) (models.QueryResult, error) {
	rows, err := db.QueryContext(ctx, statement)
	if err != nil {
		return models.QueryResult{}, err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return models.QueryResult{}, err
	}
	result := models.QueryResult{Columns: columns}
	for rows.Next() {
		values := make([]any, len(columns))
		dest := make([]any, len(columns))
		for i := range values {
			dest[i] = &values[i]
		}
		if err := rows.Scan(dest...); err != nil {
			return result, err
		}
		for i, value := range values {
			if b, ok := value.([]byte); ok {
				values[i] = string(b)
			}
		}
		result.Rows = append(result.Rows, values)
	}
	if err := rows.Err(); err != nil {
		return result, fmt.Errorf("read rows: %w", err)
	}
	result.RowCount = len(result.Rows)
	return result, nil
}

func stringPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
