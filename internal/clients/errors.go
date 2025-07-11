package clients

import (
	"errors"

	"google.golang.org/api/googleapi"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrNotFound     = errors.New("not found")

	ErrClosedChannel = errors.New("closed channel")
)

func HandleGoogleAPIError(err error) error {
	var gerr *googleapi.Error
	if errors.As(err, &gerr) {
		switch gerr.Code {
		case 401:
			return ErrUnauthorized
		case 404:
			return ErrNotFound
		}
	}

	return err
}
