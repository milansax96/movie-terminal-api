// Package service implements business logic for the API.
package service

import "errors"

// Sentinel errors returned by service methods.
var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrInvalidToken  = errors.New("invalid token")
	ErrMissingClaims = errors.New("missing required claims")
	ErrUnknownGenre  = errors.New("unknown genre")
)
