package handlers

import (
	"context"
	"net/http"

	"github.com/strax84mb/go-travel-reactive/internal/entity"
)

type authService interface {
	ValidateJwt(r *http.Request, expectedRole entity.UserRole) (string, error)
	Login(ctx context.Context, username, password string) (string, error)
}
