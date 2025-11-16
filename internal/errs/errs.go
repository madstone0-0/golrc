package errs

import "errors"

var (
	ErrNotFound  = errors.New("lyrics not found")
	ErrNotSynced = errors.New("synchronized lyrics not found")
)
