package storage

import "fmt"

type ErrPreparingStatement struct {
	query string
	cause error
}

func (e ErrPreparingStatement) Error() string {
	return fmt.Sprintf("could not prepare statement (%s) because: %s", e.query, e.cause.Error())
}

func (e ErrPreparingStatement) Unwrap() error {
	return e.cause
}

type ErrQuerying struct {
	cause error
}

func (e ErrQuerying) Error() string {
	return fmt.Sprintf("could not execute query: %s", e.cause.Error())
}

func (e ErrQuerying) Unwrap() error {
	return e.cause
}
