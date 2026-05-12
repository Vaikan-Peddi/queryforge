package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const userIDKey contextKey = "user_id"

func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	value := ctx.Value(userIDKey)
	userID, ok := value.(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return uuid.Nil, errors.New("user not authenticated")
	}
	return userID, nil
}

func UserID(r *http.Request) (uuid.UUID, error) {
	return UserIDFromContext(r.Context())
}
