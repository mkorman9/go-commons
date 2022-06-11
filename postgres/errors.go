package postgres

import (
	"errors"

	"github.com/jackc/pgconn"
	"gorm.io/gorm"
)

const (
	ErrUnknown          = iota
	ErrUniqueViolation  = iota
	ErrNotNullViolation = iota
	ErrRecordNotFound   = iota
	ErrInvalidText      = iota
)

type Error struct {
	Err        error
	Code       int
	Constraint string
	TableName  string
	ColumnName string
}

func TranslateError(err error) *pgconn.PgError {
	return err.(*pgconn.PgError)
}

func ErrorCode(err error) int {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrRecordNotFound
	}

	if pgErr, ok := err.(*pgconn.PgError); ok {
		switch pgErr.Code {
		case "23502": // not_null_violation
			return ErrNotNullViolation
		case "23505": // unique_violation
			return ErrUniqueViolation
		case "22P02": // invalid_text_representation
			return ErrInvalidText
		}
	}

	return ErrUnknown
}
