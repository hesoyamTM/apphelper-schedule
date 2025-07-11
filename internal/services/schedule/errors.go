package schedule

import "errors"

var (
	ErrUnauthorized     = errors.New("unauthorized")
	ErrScheduleNotFound = errors.New("schedule not found")
)
