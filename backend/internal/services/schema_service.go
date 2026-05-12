package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"queryforge/backend/internal/models"

	_ "modernc.org/sqlite"
)

type SchemaService struct{}

func NewSchemaService() *SchemaService {
	return &SchemaService{}
}

func (s *SchemaService) ValidateSQLiteFile(ctx context.Context, path string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	db, err := openReadOnlySQLite(path)
	if err != nil {
		return err
	}
	defer db.Close()
	var integrity string
	if err := db.QueryRowContext(ctx, `PRAGMA integrity_check`).Scan(&integrity); err != nil {
		return err
	}
	if integrity != "ok" {
		return fmt.Errorf("integrity_check failed")
	}
	return nil
}

func (s *SchemaService) Inspect(ctx context.Context, path string) (models.DatabaseSchema, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	db, err := openReadOnlySQLite(path)
	if err != nil {
		return models.DatabaseSchema{}, err
	}
	defer db.Close()

	tableRows, err := db.QueryContext(ctx, `
		SELECT name FROM sqlite_master
		WHERE type='table' AND name NOT LIKE 'sqlite_%'
		ORDER BY name
	`)
	if err != nil {
		return models.DatabaseSchema{}, err
	}
	defer tableRows.Close()

	var tableNames []string
	for tableRows.Next() {
		var tableName string
		if err := tableRows.Scan(&tableName); err != nil {
			return models.DatabaseSchema{}, err
		}
		tableNames = append(tableNames, tableName)
	}
	if err := tableRows.Err(); err != nil {
		return models.DatabaseSchema{}, err
	}
	tableRows.Close()

	var schema models.DatabaseSchema
	for _, tableName := range tableNames {
		table := models.TableSchema{Name: tableName}
		columns, err := readColumns(ctx, db, tableName)
		if err != nil {
			return schema, err
		}
		fks, err := readForeignKeys(ctx, db, tableName)
		if err != nil {
			return schema, err
		}
		count, err := rowCount(ctx, db, tableName)
		if err != nil {
			return schema, err
		}
		table.Columns = columns
		table.ForeignKeys = fks
		table.RowCount = count
		schema.Tables = append(schema.Tables, table)
	}
	return schema, nil
}

func openReadOnlySQLite(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=ro", path))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	return db, nil
}

func readColumns(ctx context.Context, db *sql.DB, table string) ([]models.ColumnSchema, error) {
	rows, err := db.QueryContext(ctx, fmt.Sprintf(`PRAGMA table_info(%q)`, table))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var columns []models.ColumnSchema
	for rows.Next() {
		var cid int
		var name, typ string
		var notNull, pk int
		var defaultValue any
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &pk); err != nil {
			return nil, err
		}
		columns = append(columns, models.ColumnSchema{Name: name, Type: typ, PrimaryKey: pk > 0, Nullable: notNull == 0})
	}
	return columns, rows.Err()
}

func readForeignKeys(ctx context.Context, db *sql.DB, table string) ([]models.ForeignKeySchema, error) {
	rows, err := db.QueryContext(ctx, fmt.Sprintf(`PRAGMA foreign_key_list(%q)`, table))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var keys []models.ForeignKeySchema
	for rows.Next() {
		var id, seq int
		var refTable, from, to, onUpdate, onDelete, match string
		if err := rows.Scan(&id, &seq, &refTable, &from, &to, &onUpdate, &onDelete, &match); err != nil {
			return nil, err
		}
		keys = append(keys, models.ForeignKeySchema{Column: from, RefTable: refTable, RefColumn: to})
	}
	return keys, rows.Err()
}

func rowCount(ctx context.Context, db *sql.DB, table string) (int64, error) {
	var count int64
	err := db.QueryRowContext(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM %q`, table)).Scan(&count)
	return count, err
}
