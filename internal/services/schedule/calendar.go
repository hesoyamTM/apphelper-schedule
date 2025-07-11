package schedule

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hesoyamTM/apphelper-schedule/internal/clients"
	"github.com/hesoyamTM/apphelper-schedule/internal/models"
	"github.com/hesoyamTM/apphelper-schedule/internal/storage"
)

type CalendarService interface {
	LoginURL(ctx context.Context, state string) string
	GetTokenFromCode(ctx context.Context, authCode string) (models.Token, error)
	GetEvents(ctx context.Context, tok models.Token, minTime, maxTime time.Time) (*[]*models.CalendarEvent, error)
	CreateEvent(ctx context.Context, tok models.Token, start, end time.Time, eventId, calendarId string) error
	DeleteEvent(ctx context.Context, tok models.Token, eventId, calendarId string) error
	CreateCalendar(ctx context.Context, tok models.Token, title string) (*models.Calendar, error)
	RefreshToken(ctx context.Context, tok models.Token) (models.Token, error)
}

type SessionStorage interface {
	SetSession(ctx context.Context, userId uuid.UUID, tok models.Token, ttl time.Duration) error
	ProvideSession(ctx context.Context, userId uuid.UUID) (models.Token, error)
	DeleteSession(ctx context.Context, userId uuid.UUID) error
}

type StateStorage interface {
	SetState(ctx context.Context, userId uuid.UUID, state string, stateTTL time.Duration) error
	GetState(ctx context.Context, userId uuid.UUID) (string, error)
	DeleteState(ctx context.Context, userId uuid.UUID) error
}

type CalendarStorage interface {
	CreateCalendar(ctx context.Context, groupId uuid.UUID, calendarId string) error
	DeleteCalendar(ctx context.Context, groupId uuid.UUID) error
	ProvideCalendar(ctx context.Context, groupId uuid.UUID) (string, error)
}

type GroupService interface {
	GetGroup(ctx context.Context, groupId uuid.UUID) (*models.Group, error)
}

type CalendarManager struct {
	sessionStorage  SessionStorage
	stateStorage    StateStorage
	calendarService CalendarService
	calendarStorage CalendarStorage
	groupService    GroupService
	stateTTL        time.Duration
}

func NewCalendarManager(
	sessionStorage SessionStorage,
	stateStorage StateStorage,
	calendarService CalendarService,
	calendarStorage CalendarStorage,
	groupService GroupService,
	stateTTL time.Duration,
) *CalendarManager {
	return &CalendarManager{
		sessionStorage:  sessionStorage,
		stateStorage:    stateStorage,
		calendarService: calendarService,
		calendarStorage: calendarStorage,
		groupService:    groupService,
		stateTTL:        stateTTL,
	}
}

func (c *CalendarManager) LoginURL(ctx context.Context, userID uuid.UUID) string {
	state := uuid.New().String()

	if err := c.stateStorage.SetState(ctx, userID, state, c.stateTTL); err != nil {
		return ""
	}

	return c.calendarService.LoginURL(ctx, state)
}

func (c *CalendarManager) Authorize(ctx context.Context, userId uuid.UUID, authcode, state string) error {
	const op = "calendar.Authorize"

	authState, err := c.stateStorage.GetState(ctx, userId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if authState != state {
		return fmt.Errorf("%s: %w", op, clients.ErrUnauthorized)
	}

	tok, err := c.calendarService.GetTokenFromCode(ctx, authcode)
	if err != nil {
		if errors.Is(err, clients.ErrUnauthorized) {
			_, err = c.refreshToken(ctx, userId, tok)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
			_, err = c.calendarService.GetTokenFromCode(ctx, authcode)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := c.stateStorage.DeleteState(ctx, userId); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := c.sessionStorage.SetSession(ctx, userId, tok, 0); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *CalendarManager) IsAuthorized(ctx context.Context, userId uuid.UUID) bool {
	_, err := c.sessionStorage.ProvideSession(ctx, userId)
	if err != nil {
		if errors.Is(err, storage.ErrSessionNotFound) {
			return false
		}

		return false
	}

	return true
}

func (c *CalendarManager) CreateEvent(ctx context.Context, userId, groupId uuid.UUID, event *models.CalendarEvent) error {
	const op = "calendar.CreateEvent"

	tok, err := c.sessionStorage.ProvideSession(ctx, userId)
	if err != nil {
		// TODO: error

		return fmt.Errorf("%s: %w", op, err)
	}

	calendarId, err := c.calendarStorage.ProvideCalendar(ctx, groupId)
	if err != nil {
		if errors.Is(err, storage.ErrCalendarNotFound) {
			calend, err := c.createCalendar(ctx, userId, groupId, tok)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
			calendarId = calend.Id
		} else {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	if err := c.calendarService.CreateEvent(ctx, tok, event.Start, event.End, event.Title, calendarId); err != nil {
		if errors.Is(err, clients.ErrUnauthorized) {
			tok, err = c.refreshToken(ctx, userId, tok)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
			if err := c.calendarService.CreateEvent(ctx, tok, event.Start, event.End, event.Title, calendarId); err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
			return nil
		}
		if errors.Is(err, clients.ErrNotFound) {
			calend, err := c.createCalendar(ctx, userId, groupId, tok)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}

			if err = c.calendarService.CreateEvent(ctx, tok, event.Start, event.End, event.Title, calend.Id); err != nil && !errors.Is(err, clients.ErrNotFound) {
				return fmt.Errorf("%s: %w", op, err)
			}
			return nil
		}

		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (c *CalendarManager) GetEvents(ctx context.Context, userId uuid.UUID, minTime, maxTime time.Time) (*[]*models.CalendarEvent, error) {
	const op = "calendar.GetEvents"

	tok, err := c.sessionStorage.ProvideSession(ctx, userId)
	if err != nil {
		// TODO: error

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	events, err := c.calendarService.GetEvents(ctx, tok, minTime, maxTime)
	if err != nil {
		if errors.Is(err, clients.ErrUnauthorized) {
			tok, err = c.refreshToken(ctx, userId, tok)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", op, err)
			}
			events, err = c.calendarService.GetEvents(ctx, tok, minTime, maxTime)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", op, err)
			}
			return events, nil
		}
		// TODO: error

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return events, nil
}

func (c *CalendarManager) DeleteEvent(ctx context.Context, userId uuid.UUID, group_id uuid.UUID, eventId string) error {
	const op = "calendar.DeleteEvent"

	tok, err := c.sessionStorage.ProvideSession(ctx, userId)
	if err != nil {
		// TODO: error

		return fmt.Errorf("%s: %w", op, err)
	}

	calendarId, err := c.calendarStorage.ProvideCalendar(ctx, group_id)
	if err != nil {
		// TODO: error

		return fmt.Errorf("%s: %w", op, err)
	}

	if err := c.calendarService.DeleteEvent(ctx, tok, eventId, calendarId); err != nil {
		if errors.Is(err, clients.ErrUnauthorized) {
			tok, err = c.refreshToken(ctx, userId, tok)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
			if err := c.calendarService.DeleteEvent(ctx, tok, eventId, calendarId); err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
		}
		if errors.Is(err, clients.ErrNotFound) {
			return nil
		}
		// TODO: error

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *CalendarManager) createCalendar(ctx context.Context, userId, groupId uuid.UUID, tok models.Token) (*models.Calendar, error) {
	const op = "calendar.createCalendar"

	group, err := c.groupService.GetGroup(ctx, groupId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	calend, err := c.calendarService.CreateCalendar(ctx, tok, group.Name)
	if err != nil {
		if errors.Is(err, clients.ErrUnauthorized) {
			tok, err = c.refreshToken(ctx, userId, tok)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", op, err)
			}
			calend, err = c.calendarService.CreateCalendar(ctx, tok, group.Name)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", op, err)
			}
		} else {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	if err := c.calendarStorage.CreateCalendar(ctx, groupId, calend.Id); err != nil {
		// TODO: error

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return calend, nil
}

func (c *CalendarManager) refreshToken(ctx context.Context, userId uuid.UUID, tok models.Token) (models.Token, error) {
	const op = "calendar.RefreshToken"

	newTok, err := c.calendarService.RefreshToken(ctx, tok)
	if err != nil {
		if errors.Is(err, clients.ErrUnauthorized) {
			// if err := c.sessionStorage.DeleteSession(ctx, userId); err != nil {
			// 	return models.Token{}, fmt.Errorf("%s: %w", op, err)
			// }
			return models.Token{}, fmt.Errorf("%s: %w", op, err)
		}

		return models.Token{}, fmt.Errorf("%s: %w", op, err)
	}

	if err := c.sessionStorage.SetSession(ctx, userId, newTok, 0); err != nil {
		return models.Token{}, fmt.Errorf("%s: %w", op, err)
	}

	return newTok, nil
}
