package errors

import (
	"errors"
	"fmt"
)

// Erreurs métier standardisées pour toute l'API
var (
	ErrBadRequest      = errors.New("bad_request")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrForbidden       = errors.New("forbidden")
	ErrNotFound        = errors.New("not_found")
	ErrConflict        = errors.New("conflict")
	ErrTooManyRequests = errors.New("too_many_requests")
	ErrInternalError   = errors.New("internal_error")
)

// Wrappers pratiques pour créer des erreurs typées
func NewBadRequest(msg string) error {
	return fmt.Errorf("%w: %s", ErrBadRequest, msg)
}

func NewUnauthorized(msg string) error {
	return fmt.Errorf("%w: %s", ErrUnauthorized, msg)
}

func NewForbidden(msg string) error {
	return fmt.Errorf("%w: %s", ErrForbidden, msg)
}

func NewNotFound(msg string) error {
	return fmt.Errorf("%w: %s", ErrNotFound, msg)
}

func NewConflict(msg string) error {
	return fmt.Errorf("%w: %s", ErrConflict, msg)
}

func NewInternalError(msg string) error {
	return fmt.Errorf("%w: %s", ErrInternalError, msg)
}
