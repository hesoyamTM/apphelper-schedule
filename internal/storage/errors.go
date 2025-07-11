package storage

import "errors"

var (
	ErrSessionNotFound  = errors.New("session not found")
	ErrScheduleNotFound = errors.New("schedule not found")
	ErrGroupNotFound    = errors.New("group not found")
	ErrEventNotFound    = errors.New("event not found")
	ErrInvalidUUID      = errors.New("invalid uuid")
	ErrCalendarNotFound = errors.New("calendar not found")
	ErrStateNotFound    = errors.New("state not found")
)
