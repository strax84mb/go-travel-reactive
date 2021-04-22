package auth

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/reactivex/rxgo/v2"

	"github.com/dgrijalva/jwt-go"
	"github.com/strax84mb/go-travel-reactive/internal/app"
	"github.com/strax84mb/go-travel-reactive/internal/entity"
)

type repository interface {
	GetUserByUsername(username string) (entity.User, error)
	SaveUser(user entity.User) (int, error)
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

func (a *authService) generateJwt() func(context.Context, interface{}) (interface{}, error) {
	return func(ctx context.Context, item interface{}) (interface{}, error) {
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
}

func (a *authService) validRole(expectedRole, role, roleFromDb entity.UserRole) bool {
	if expectedRole == "ANY" {
		return true
	}
	return expectedRole == role && role == roleFromDb
}

// ValidateJwt returns username
func (a *authService) ValidateJwt(r *http.Request, expectedRole entity.UserRole) (string, error) {
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

		user, err := a.repo.GetUserByUsername(username)
		if err != nil {
			return []byte{}, fmt.Errorf("can't load user data: %w", err)
		}

		roleFromDb = user.Role

		return user.Salt, nil
	})
	if err != nil {
		return "", fmt.Errorf("error while parsing JWT: %w", err)
	}

	claims := token.Claims.(jwt.MapClaims)
	username := claims["sub"].(string)
	role := entity.UserRole(claims["role"].(string))

	if !a.validRole(expectedRole, role, roleFromDb) {
		return "", errors.New("incorrect role")
	}
	return username, nil
}

func (a *authService) getUserByUsername() func(context.Context, interface{}) (interface{}, error) {
	return func(ctx context.Context, item interface{}) (interface{}, error) {
		user, err := a.repo.GetUserByUsername(item.(string))
		if err != nil {
			err = fmt.Errorf("can't load user with username %s: %w", item.(string), err)
		}

		return user, err
	}
}

func (a *authService) validatePassword(password string) func(context.Context, interface{}) (interface{}, error) {
	return func(ctx context.Context, item interface{}) (interface{}, error) {
		user := item.(entity.User)
		encodedPassword := encodePassword(password, user.Salt)

		if encodedPassword != user.Password {
			return "", errors.New("wrong password")
		}

		return item, nil
	}
}

// Login returns JWT
func (a *authService) Login(ctx context.Context, username, password string) (string, error) {
	item := <-rxgo.JustItem(username).
		Map(a.getUserByUsername()).
		Map(a.validatePassword(password)).
		Map(a.generateJwt()).
		Observe()

	if item.E != nil {
		a.logger.Error(app.ContextWithError(ctx, item.E), "login failed for username %s", username)
		return "", fmt.Errorf("login failed: %w", item.E)
	}

	return item.V.(string), nil
}

func encodePassword(password string, salt []byte) string {
	h := sha512.New()
	_, _ = h.Write(salt)
	_, _ = h.Write([]byte(password))
	hashedPassword := h.Sum(salt)
	return hex.EncodeToString(hashedPassword)
}

func generateSalt() []byte {
	salt := make([]byte, 16)
	rand.Seed(time.Now().UnixNano())
	rand.Read(salt)
	return salt
}

type usernameAndPassword struct {
	Username string
	Password string
}

func (a *authService) checkIfUserExists(username string) func(context.Context, interface{}) (interface{}, error) {
	return func(ctx context.Context, item interface{}) (interface{}, error) {
		if _, err := a.repo.GetUserByUsername(username); err != nil {
			var notFound entity.ErrUsernameNotFound
			if !errors.As(err, &notFound) {
				return item, nil
			}

			return item, err
		}

		return item, entity.ErrUsernameTaken{Username: username}
	}
}

func (a *authService) SaveUser(ctx context.Context, username, password string) (int, error) {
	item := <-rxgo.JustItem(usernameAndPassword{
		Username: username,
		Password: password,
	}).Map(a.checkIfUserExists(username)).Map(func(ctx context.Context, item interface{}) (interface{}, error) {
		unp := item.(usernameAndPassword)
		user := entity.User{
			Username: unp.Username,
			Role:     entity.CommonUserRole,
		}
		user.Salt = generateSalt()
		user.Password = encodePassword(user.Password, user.Salt)

		id, err := a.repo.SaveUser(user)
		if err != nil {
			return 0, fmt.Errorf("could not save user: %w", err)
		}

		return id, nil
	}).Observe()
	if item.E != nil {
		a.logger.Error(app.ContextWithError(ctx, item.E), "failed to save new user")
		return 0, fmt.Errorf("failed to save new user: %w", item.E)
	}

	return item.V.(int), nil
}
