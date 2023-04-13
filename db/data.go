package db

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

// UNIQUE constraint failed
func IsUniqueError(err error) bool {
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}

// Nodata failed
func IsNoDataError(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
