package storage

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/strax84mb/go-travel-reactive/internal/entity"
)

func NewRepository(dsn string) (*repository, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("could not connect: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return &repository{db: db}, nil
}

type repository struct {
	db *sql.DB
}

// GetUserByUsername input is username
func (r *repository) GetUserByUsername(ctx context.Context, usernameItem interface{}) (interface{}, error) {
	username := usernameItem.(string)
	query := `SELECT id, password, salt, role FROM users WHERE username = ?`

	stmt, err := r.db.Prepare(query)
	if err != nil {
		return entity.User{}, makeErrPreparingStatement(query, err)
	}

	defer stmt.Close()

	var salt string

	user := entity.User{
		Username: username,
	}

	err = stmt.QueryRowContext(ctx, username).Scan(&user.ID, &user.Password, &salt, &user.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.User{}, entity.ErrUsernameNotFound
		}

		return entity.User{}, ErrQuerying{cause: err}
	}

	if user.Salt, err = hex.DecodeString(salt); err != nil {
		return entity.User{}, fmt.Errorf("can't decode salt: %w", err)
	}

	return user, nil

}

// SaveUser returns last inserted ID
// input is entity.User
func (r *repository) SaveUser(ctx context.Context, userItem interface{}) (interface{}, error) {
	user := userItem.(entity.User)
	query := `INSERT INTO users (username, password, salt, role) VALUES (?, ?, ?, ?)`

	stmt, err := r.db.Prepare(query)
	if err != nil {
		return 0, makeErrPreparingStatement(query, err)
	}

	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, user.Username, user.Password, hex.EncodeToString(user.Salt), user.Role)
	if err != nil {
		return 0, ErrQuerying{cause: err}
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("could not get last inserted ID: %w", err)
	}

	return int(id), nil
}
