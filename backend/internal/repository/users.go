package repository

import (
	"context"
	"errors"

	"queryforge/backend/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, name, email, passwordHash string) (models.User, error) {
	var user models.User
	err := r.pool.QueryRow(ctx, `
		INSERT INTO users(name, email, password_hash)
		VALUES($1, lower($2), $3)
		RETURNING id, name, email, password_hash, created_at
	`, name, email, passwordHash).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt)
	return user, err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, email, password_hash, created_at FROM users WHERE email=lower($1)
	`, email).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return user, ErrNotFound
	}
	return user, err
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (models.User, error) {
	var user models.User
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, email, password_hash, created_at FROM users WHERE id=$1
	`, id).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return user, ErrNotFound
	}
	return user, err
}
