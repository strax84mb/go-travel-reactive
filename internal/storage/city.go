package storage

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/strax84mb/go-travel-reactive/internal/entity"
)

func (r *repository) GetCityByNameAndCountry(name, country string) (entity.City, error) {
	query := `SELECT id, name, country FROM cities WHERE LOWER(name) = LOWER(?) AND LOWER(country) = LOWER(?)`

	stmt, err := r.db.Prepare(query)
	if err != nil {
		return entity.City{}, makeErrPreparingStatement(query, err)
	}

	city := entity.City{}

	err = stmt.QueryRow(name, country).Scan(&city.ID, &city.Name, &city.Country)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.City{}, entity.ErrCityNotFound
		}

		return entity.City{}, ErrQuerying{cause: err}
	}

	return city, nil
}

func (r *repository) AddCity(name, country string) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, ErrBeginTx{cause: err}
	}

	query := `SELECT count(id) FROM cities WHERE LOWER(name) = LOWER(?) AND LOWER(country) = LOWER(?)`

	checkStmt, err := tx.Prepare(query)
	if err != nil {
		return 0, makeErrPreparingStatement(query, err)
	}

	defer checkStmt.Close()

	var count int

	if err = checkStmt.QueryRow(name, country).Scan(&count); err != nil {
		return 0, ErrQuerying{cause: err}
	}

	stmt, err := tx.Prepare(`INSERT INTO cities (name, country) VALUES (?, ?)`)
	if err != nil {
		return 0, makeErrPreparingStatement(`INSERT INTO users (name, country) VALUES (?, ?)`, err)
	}

	defer stmt.Close()

	result, err := stmt.Exec(name, country)
	if err != nil {
		return 0, ErrQuerying{cause: err}
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("could not get last inserted ID: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return 0, ErrCommitTx{cause: err}
	}

	return int(id), nil
}

func (r *repository) UpdateCity(city entity.City) error {
	statement := `UPDATE cities SET name=?, country=? WHERE id=?`

	stmt, err := r.db.Prepare(statement)
	if err != nil {
		return makeErrPreparingStatement(statement, err)
	}

	result, err := stmt.Exec(city.Name, city.Country, city.ID)
	if err != nil {
		return ErrQuerying{cause: err}
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not get number of affected rows: %w", err)
	} else if affected == 0 {
		return entity.ErrCityNotFound
	}

	return nil
}

func (r *repository) GetCity(id int) (entity.City, error) {
	query := `SELECT name, country FROM cities WHERE id=?`

	stmt, err := r.db.Prepare(query)
	if err != nil {
		return entity.City{}, makeErrPreparingStatement(query, err)
	}

	defer stmt.Close()

	city := entity.City{ID: id}
	if err = stmt.QueryRow(id).Scan(&city.Name, &city.Country); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.City{}, entity.ErrCityNotFound
		}

		return entity.City{}, ErrScanning{cause: err}
	}

	return city, nil
}

func (r *repository) GetAllCities() ([]entity.City, error) {
	query := `SELECT id, name, country FROM cities`

	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, makeErrPreparingStatement(query, err)
	}

	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return nil, ErrQuerying{cause: err}
	}

	var (
		result []entity.City
		city   entity.City
	)

	for rows.Next() {
		if err = rows.Scan(&city.ID, &city.Name, &city.Country); err != nil {
			return nil, ErrScanning{cause: err}
		}

		result = append(result, city)

		if rows.Err() != nil {
			return nil, ErrIteration{cause: err}
		}
	}

	return result, nil
}
