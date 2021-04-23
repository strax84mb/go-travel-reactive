package handlers

import (
	"context"
	"net/http"

	"github.com/strax84mb/go-travel-reactive/internal/entity"
)

type authService interface {
	ValidateJwt(ctx context.Context, r *http.Request, expectedRole entity.UserRole) (string, error)
	Login(ctx context.Context, username, password string) (string, error)
	SaveUser(ctx context.Context, username, password string) (int, error)
}
