package pq

import (
	"github.com/jackc/pgconn"
)

const uniqueViolationCode = "23505"

func IsDuplicated(err error) bool {
	// unique_violation
	if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == uniqueViolationCode {
		return true
	}

	return false
}
