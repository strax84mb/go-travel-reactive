package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/strax84mb/go-travel-reactive/internal/entity"
)

func (r *repository) GetCityByNameAndCountry(ctx context.Context, cityItem interface{}) (interface{}, error) {
	city := cityItem.(entity.City)
	query := `SELECT id, name, country FROM cities WHERE LOWER(name) = LOWER(?) AND LOWER(country) = LOWER(?)`

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return entity.City{}, makeErrPreparingStatement(query, err)
	}

	result := entity.City{}

	err = stmt.QueryRowContext(ctx, city.Name, city.Country).Scan(&result.ID, &result.Name, &result.Country)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.City{}, entity.ErrCityNotFound
		}

		return entity.City{}, ErrQuerying{cause: err}
	}

	return result, nil
}

func (r *repository) AddCity(ctx context.Context, cityItem interface{}) (interface{}, error) {
	city := cityItem.(entity.City)

	tx, err := r.db.Begin()
	if err != nil {
		return 0, ErrBeginTx{cause: err}
	}

	query := `SELECT count(id) FROM cities WHERE LOWER(name) = LOWER(?) AND LOWER(country) = LOWER(?)`

	checkStmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return 0, makeErrPreparingStatement(query, err)
	}

	defer checkStmt.Close()

	var count int

	if err = checkStmt.QueryRowContext(ctx, city.Name, city.Country).Scan(&count); err != nil {
		return 0, ErrQuerying{cause: err}
	}

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO cities (name, country) VALUES (?, ?)`)
	if err != nil {
		return 0, makeErrPreparingStatement(`INSERT INTO users (name, country) VALUES (?, ?)`, err)
	}

	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, city.Name, city.Country)
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

func (r *repository) UpdateCity(ctx context.Context, cityItem interface{}) (interface{}, error) {
	city := cityItem.(entity.City)
	statement := `UPDATE cities SET name=?, country=? WHERE id=?`

	stmt, err := r.db.PrepareContext(ctx, statement)
	if err != nil {
		return false, makeErrPreparingStatement(statement, err)
	}

	result, err := stmt.ExecContext(ctx, city.Name, city.Country, city.ID)
	if err != nil {
		return false, ErrQuerying{cause: err}
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("could not get number of affected rows: %w", err)
	} else if affected == 0 {
		return false, entity.ErrCityNotFound
	}

	return true, nil
}

func (r *repository) GetCity(ctx context.Context, id interface{}) (interface{}, error) {
	query := `SELECT name, country FROM cities WHERE id=?`

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return entity.City{}, makeErrPreparingStatement(query, err)
	}

	defer stmt.Close()

	city := entity.City{ID: id.(int)}
	if err = stmt.QueryRowContext(ctx, id).Scan(&city.Name, &city.Country); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.City{}, entity.ErrCityNotFound
		}

		return entity.City{}, ErrScanning{cause: err}
	}

	return city, nil
}

func (r *repository) GetAllCities(ctx context.Context, _ interface{}) (interface{}, error) {
	query := `SELECT id, name, country FROM cities`

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, makeErrPreparingStatement(query, err)
	}

	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx)
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

func (r *repository) DeleteCity(ctx context.Context, idItem interface{}) (interface{}, error) {
	id := idItem.(int)

	tx, err := r.db.Begin()
	if err != nil {
		return 0, ErrBeginTx{cause: err}
	}

	// delete routes
	query := `DELETE FROM routes WHERE 
		source_id IN (SELECT id FROM airports WHERE city_id=?) OR 
		destination_id IN (SELECT id FROM airports WHERE city_id=?)`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return 0, makeErrPreparingStatement(query, err)
	}

	if _, err = stmt.ExecContext(ctx, id, id); err != nil {
		_ = stmt.Close()
		return 0, ErrQuerying{cause: err}
	}

	// delete airports
	_ = stmt.Close()
	query = `DELETE FROM airports WHERE city_id=?`

	stmt, err = tx.PrepareContext(ctx, query)
	if err != nil {
		_ = tx.Rollback()
		return 0, makeErrPreparingStatement(query, err)
	}

	if _, err = stmt.ExecContext(ctx, id); err != nil {
		_, _ = stmt.Close(), tx.Rollback()
		return 0, ErrQuerying{cause: err}
	}

	// delete comments
	_ = stmt.Close()
	query = `DELETE FROM comments WHERE city_id=?`

	stmt, err = tx.PrepareContext(ctx, query)
	if err != nil {
		_ = tx.Rollback()
		return 0, makeErrPreparingStatement(query, err)
	}

	if _, err = stmt.ExecContext(ctx, id); err != nil {
		_, _ = stmt.Close(), tx.Rollback()
		return 0, ErrQuerying{cause: err}
	}

	// delete city
	_ = stmt.Close()
	query = `DELETE FROM cities WHERE id=?`

	stmt, err = tx.PrepareContext(ctx, query)
	if err != nil {
		_ = tx.Rollback()
		return 0, makeErrPreparingStatement(query, err)
	}

	result, err := stmt.ExecContext(ctx, id)
	if err != nil {
		_, _ = stmt.Close(), tx.Rollback()
		return 0, ErrQuerying{cause: err}
	}

	_ = stmt.Close()

	count, err := result.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return 0, fmt.Errorf("could not get number of affected rows: %w", err)
	} else if count == 0 {
		_ = tx.Rollback()
		return 0, entity.ErrCityNotFound
	}

	if err = tx.Commit(); err != nil {
		_, _ = stmt.Close(), tx.Rollback()
		return 0, ErrCommitTx{cause: err}
	}

	return count, nil
}
