package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/reactivex/rxgo/v2"
	"github.com/strax84mb/go-travel-reactive/internal/app"
	"github.com/strax84mb/go-travel-reactive/internal/entity"
)

type repository interface {
	GetUserByUsername(ctx context.Context, username interface{}) (interface{}, error)
	SaveUser(ctx context.Context, user interface{}) (interface{}, error)
}

type authService struct {
	repo   repository
	logger app.Logger
}

func NewAuthService(repo repository, logger app.Logger) *authService {
	return &authService{
		repo:   repo,
		logger: logger,
	}
}

func generateJwt(ctx context.Context, item interface{}) (interface{}, error) {
	user := item.(entity.User)
	now := time.Now()
	exp := now.Add(3600 * 1000000000)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.Username,
		"role": user.Role,
		"nbf":  now.Unix(),
		"iat":  now.Unix(),
		"exp":  exp.Unix(),
	})

	tokenString, err := token.SignedString(user.Salt)
	if err != nil {
		return "", fmt.Errorf("failed to sign jwt token: %w", err)
	}

	return tokenString, nil
}

// ValidateJwt returns username
func (a *authService) ValidateJwt(ctx context.Context, r *http.Request, expectedRole entity.UserRole) (string, error) {
	header := r.Header.Get("Authorization")

	if header == "" {
		return "", errors.New("authorization header is missing")
	}

	if !strings.HasPrefix(header, "Bearer ") {
		return "", errors.New("missing authentication token")
	}

	var roleFromDb entity.UserRole

	token, err := jwt.Parse(strings.TrimPrefix(header, "Bearer "), func(token *jwt.Token) (interface{}, error) {
		claims := token.Claims.(jwt.MapClaims)
		username := claims["sub"].(string)

		userItem, err := a.repo.GetUserByUsername(ctx, username)
		if err != nil {
			return []byte{}, fmt.Errorf("can't load user data: %w", err)
		}

		user := userItem.(entity.User)
		roleFromDb = user.Role

		return user.Salt, nil
	})
	if err != nil {
		return "", fmt.Errorf("error while parsing JWT: %w", err)
	}

	claims := token.Claims.(jwt.MapClaims)
	username := claims["sub"].(string)
	role := entity.UserRole(claims["role"].(string))

	if !validRole(expectedRole, role, roleFromDb) {
		return "", errors.New("incorrect role")
	}
	return username, nil
}

func validatePassword(ctx context.Context, item interface{}) (interface{}, error) {
	user := item.(entity.User)
	encodedPassword := encodePassword(ctx.Value(ctxPasswordIdx).(string), user.Salt)

	if encodedPassword != user.Password {
		return "", errors.New("wrong password")
	}

	return item, nil
}

// Login returns JWT
func (a *authService) Login(ctx context.Context, username, password string) (string, error) {
	ctx = app.ContextWithValue(ctx, "function", "authService.Login")

	item := <-rxgo.JustItem(username).
		Map(a.repo.GetUserByUsername).
		Map(validatePassword, rxgo.WithContext(context.WithValue(ctx, ctxPasswordIdx, password))).
		Map(generateJwt).
		Observe()
	if item.Error() {
		a.logger.Error(app.ContextWithError(ctx, item.E), "login failed for username %s", username)
		return "", fmt.Errorf("login failed: %w", item.E)
	}

	return item.V.(string), nil
}

type usernameAndPassword struct {
	Username string
	Password string
}

// Should return boolean
func checkIfUserExists(ctx context.Context, item interface{}) (interface{}, error) {
	if err, ok := item.(error); ok {
		if errors.Is(err, entity.ErrUsernameNotFound) {
			return false, nil
		}

		return false, err
	}

	return true, entity.ErrUsernameTaken
}

func createAndEncodeUser(_ context.Context, i interface{}, item interface{}) (interface{}, error) {
	unp := item.(usernameAndPassword)
	user := entity.User{
		Username: unp.Username,
		Role:     entity.CommonUserRole,
	}
	user.Salt = generateSalt()
	user.Password = encodePassword(unp.Password, user.Salt)

	return user, nil
}

func currentTime(_ interface{}) time.Time {
	return time.Now()
}

func (a *authService) SaveUser(ctx context.Context, username, password string) (int, error) {
	ctx = app.ContextWithValue(ctx, "function", "authService.SaveUser")
	ctx = app.ContextWithValue(ctx, "username", username)
	unp := usernameAndPassword{
		Username: username,
		Password: password,
	}

	item := <-rxgo.Just(username)().
		Map(a.repo.GetUserByUsername).
		OnErrorReturn(func(err error) interface{} {
			return err
		}).
		Map(checkIfUserExists).
		Join(createAndEncodeUser, rxgo.Just(unp)(), currentTime, rxgo.WithDuration(2*time.Second)).
		Map(a.repo.SaveUser).
		Observe()
	if item.Error() {
		a.logger.Error(app.ContextWithError(ctx, item.E), "failed to save new user")
		return 0, fmt.Errorf("failed to save new user: %w", item.E)
	}

	return item.V.(int), nil
}
