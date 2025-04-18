package internal

import (
	"context"
	"server/models"
)

type contextKey string

const (
	userIDContextKey contextKey = "userID"
	roleContextKey   contextKey = "role"
)

// SetUserContext adds user ID and role to the request context
func SetUserContext(ctx context.Context, userID int) context.Context {
	ctx = context.WithValue(ctx, userIDContextKey, userID)
	return ctx
}

// GetUserIDFromContext extracts the user ID from the context
func GetUserIDFromContext(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(userIDContextKey).(int)
	return userID, ok
}

// GetRoleFromContext extracts the user role from the context
func GetRoleFromContext(ctx context.Context) (models.Role, bool) {
	role, ok := ctx.Value(roleContextKey).(models.Role)
	return role, ok
}
