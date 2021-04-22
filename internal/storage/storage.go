package storage

import (
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

func (r *repository) GetUserByUsername(username string) (entity.User, error) {
	query := `SELECT id, password, salt, role FROM users WHERE username = ?`

	stmt, err := r.db.Prepare(query)
	if err != nil {
		return entity.User{}, ErrPreparingStatement{
			query: query,
			cause: err,
		}
	}

	defer stmt.Close()

	var salt string

	user := entity.User{
		Username: username,
	}

	if err := stmt.QueryRow(username).Scan(&user.ID, &user.Password, &salt, &user.Role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.User{}, entity.ErrUsernameNotFound{Username: username}
		}

		return entity.User{}, ErrQuerying{cause: err}
	}

	if user.Salt, err = hex.DecodeString(salt); err != nil {
		return entity.User{}, fmt.Errorf("can't decode salt: %w", err)
	}

	return user, nil
}

// SaveUser returns last inserted ID
func (r *repository) SaveUser(user entity.User) (int, error) {
	query := `INSERT INTO users (username, password, salt, role) VALUES (?, ?, ?, ?)`

	stmt, err := r.db.Prepare(query)
	if err != nil {
		return 0, ErrPreparingStatement{
			query: query,
			cause: err,
		}
	}

	defer stmt.Close()

	result, err := stmt.Exec(user.Username, user.Password, hex.EncodeToString(user.Salt), user.Role)
	if err != nil {
		return 0, ErrQuerying{cause: err}
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("could not get last inserted ID: %w", err)
	}

	return int(id), nil
}
