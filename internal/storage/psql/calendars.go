package psql

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/hesoyamTM/apphelper-schedule/internal/storage"
	"github.com/jackc/pgx/v5"
)

func (s *Storage) CreateCalendar(ctx context.Context, groupId uuid.UUID, calendarId string) error {
	const op = "psql.CreateCalendar"

	query := `INSERT INTO calendars (group_id, calendar_id)
	VALUES ($1, $2)`

	if _, err := s.db.Exec(ctx, query, groupId, calendarId); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) DeleteCalendar(ctx context.Context, groupId uuid.UUID) error {
	const op = "psql.DeleteCalendar"

	query := `DELETE FROM calendars WHERE group_id = $1`

	if _, err := s.db.Exec(ctx, query, groupId); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) ProvideCalendar(ctx context.Context, groupId uuid.UUID) (string, error) {
	const op = "psql.ProvideCalendar"

	query := `SELECT calendar_id FROM calendars WHERE group_id = $1`

	row := s.db.QueryRow(ctx, query, groupId)

	var calendarId string

	if err := row.Scan(&calendarId); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", storage.ErrCalendarNotFound
		}

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return calendarId, nil
}
