package store

import "errors"

var (
	ErrActivityNotFound    = errors.New("activity_not_found")
	ErrRegistrationClosed  = errors.New("registration_closed")
	ErrCancellationClosed  = errors.New("cancellation_closed")
	ErrAlreadyRegistered   = errors.New("already_registered")
	ErrNotRegistered       = errors.New("not_registered")
	ErrActivityTimeInvalid = errors.New("activity_time_invalid")
	ErrForbidden           = errors.New("forbidden")
)
