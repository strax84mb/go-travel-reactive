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

func makeErrPreparingStatement(query string, cause error) error {
	return ErrPreparingStatement{
		query: query,
		cause: cause,
	}
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

type ErrScanning struct {
	cause error
}

func (e ErrScanning) Error() string {
	return fmt.Sprintf("could not scan values: %s", e.cause.Error())
}

func (e ErrScanning) Unwrap() error {
	return e.cause
}

type ErrBeginTx struct {
	cause error
}

func (e ErrBeginTx) Error() string {
	return fmt.Sprintf("could not begin transaction: %s", e.cause.Error())
}

func (e ErrBeginTx) Unwrap() error {
	return e.cause
}

type ErrCommitTx struct {
	cause error
}

func (e ErrCommitTx) Error() string {
	return fmt.Sprintf("could not commit transaction: %s", e.cause.Error())
}

func (e ErrCommitTx) Unwrap() error {
	return e.cause
}

type ErrIteration struct {
	cause error
}

func (e ErrIteration) Error() string {
	return fmt.Sprintf("could not iterate through rows: %s", e.cause.Error())
}

func (e ErrIteration) Unwrap() error {
	return e.cause
}
