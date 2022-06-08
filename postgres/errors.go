package postgres

import (
	"errors"

	"github.com/jackc/pgconn"
	"gorm.io/gorm"
)

const (
	ErrUniqueViolation  = iota
	ErrNotNullViolation = iota
	ErrRecordNotFound   = iota
	ErrInvalidValue     = iota
)

type Error struct {
	Err        error
	Code       int
	Constraint string
}

func TranslateError(err error) (*Error, bool) {
	postgresError := Error{Err: err, Code: -1}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		postgresError.Code = ErrRecordNotFound
		return &postgresError, true
	}

	if pgErr, ok := err.(*pgconn.PgError); ok {
		switch pgErr.Code {
		case "23502": // not_null_violation
			postgresError.Code = ErrNotNullViolation
			postgresError.Constraint = pgErr.ConstraintName
		case "23505": // unique_violation
			postgresError.Code = ErrUniqueViolation
			postgresError.Constraint = pgErr.ConstraintName
		case "22P02": // invalid_text_representation
			postgresError.Code = ErrInvalidValue
		}
	}

	if postgresError.Code != -1 {
		return &postgresError, true
	} else {
		return nil, false
	}
}
