package stats

import (
	"errors"
)

var (
	// ErrNoSuchSlice is returned when there is no slice for the
	// given key.
	ErrNoSuchSlice = errors.New("no slice exists for that key")
)
