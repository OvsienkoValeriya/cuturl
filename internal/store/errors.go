package store

import (
	"errors"
)

var ErrUniqueViolation = errors.New("unique violation")
