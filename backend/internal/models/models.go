package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type Workspace struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	Name           string    `json:"name"`
	DBType         string    `json:"db_type"`
	SQLiteFilePath string    `json:"-"`
	HasDatabase    bool      `json:"has_database"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type QueryHistory struct {
	ID           uuid.UUID `json:"id"`
	WorkspaceID  uuid.UUID `json:"workspace_id"`
	UserID       uuid.UUID `json:"user_id"`
	Question     *string   `json:"question,omitempty"`
	GeneratedSQL *string   `json:"generated_sql,omitempty"`
	ExecutedSQL  *string   `json:"executed_sql,omitempty"`
	Explanation  *string   `json:"explanation,omitempty"`
	Status       string    `json:"status"`
	ErrorMessage *string   `json:"error_message,omitempty"`
	ExecutionMS  *int64    `json:"execution_ms,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type ColumnSchema struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	PrimaryKey bool   `json:"primary_key"`
	Nullable   bool   `json:"nullable"`
}

type ForeignKeySchema struct {
	Column    string `json:"column"`
	RefTable  string `json:"ref_table"`
	RefColumn string `json:"ref_column"`
}

type TableSchema struct {
	Name        string             `json:"name"`
	Columns     []ColumnSchema     `json:"columns"`
	ForeignKeys []ForeignKeySchema `json:"foreign_keys"`
	RowCount    int64              `json:"row_count"`
}

type DatabaseSchema struct {
	Tables []TableSchema `json:"tables"`
}

type QueryResult struct {
	Columns     []string `json:"columns"`
	Rows        [][]any  `json:"rows"`
	RowCount    int      `json:"row_count"`
	ExecutionMS int64    `json:"execution_ms"`
}
