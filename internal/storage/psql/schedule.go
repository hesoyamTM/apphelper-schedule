package psql

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hesoyamTM/apphelper-schedule/internal/models"
	"github.com/hesoyamTM/apphelper-schedule/internal/storage"
	"github.com/hesoyamTM/apphelper-sso/pkg/logger"
	"github.com/jackc/pgx/v5"
)

func (s *Storage) CreateSchedule(ctx context.Context, sched *models.Schedule) error {
	const op = "psql.CreateSchedule"

	query := `INSERT INTO schedules (group_id, title, student_id, trainer_id, start_date, end_date)
	VALUES ($1, $2, $3, $4, $5, $6)`

	if _, err := s.db.Exec(ctx, query, sched.GroupId, sched.Title, sched.StudentId, sched.TrainerId, sched.Start, sched.End); err != nil {
		// TODO: error

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) ProvideSchedules(ctx context.Context, trainerId, studentId uuid.UUID) ([]*models.Schedule, error) {
	if trainerId != uuid.Nil && studentId != uuid.Nil {
		query := `SELECT groups.name, schedules.group_id, schedules.title, schedules.student_id, schedules.trainer_id, schedules.start_date, schedules.end_date
		FROM schedules
		INNER JOIN groups ON groups.id = schedules.group_id
		WHERE schedules.trainer_id = $1 AND schedules.student_id = $2`
		return s.provideSchedules(ctx, func() (pgx.Rows, error) { return s.db.Query(ctx, query, trainerId, studentId) })
	}
	if trainerId != uuid.Nil {
		query := `SELECT groups.name, schedules.group_id, schedules.title, schedules.student_id, schedules.trainer_id, schedules.start_date, schedules.end_date
		FROM schedules
		INNER JOIN groups ON groups.id = schedules.group_id
		WHERE schedules.trainer_id = $1`
		return s.provideSchedules(ctx, func() (pgx.Rows, error) { return s.db.Query(ctx, query, trainerId) })
	}
	if studentId != uuid.Nil {
		query := `SELECT groups.name, schedules.group_id, schedules.title, schedules.student_id, schedules.trainer_id, schedules.start_date, schedules.end_date
		FROM schedules
		INNER JOIN groups ON groups.id = schedules.group_id
		WHERE schedules.student_id = $1`
		return s.provideSchedules(ctx, func() (pgx.Rows, error) { return s.db.Query(ctx, query, studentId) })
	}
	return nil, nil
}

func (s *Storage) provideSchedules(ctx context.Context, queryFunc func() (pgx.Rows, error)) ([]*models.Schedule, error) {
	const op = "psql.provideSchedules"

	rows, err := queryFunc()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	schedules := make([]*models.Schedule, 0)
	for rows.Next() {
		var schedule models.Schedule
		var groupId, studentId, trainerId uuid.NullUUID

		if err := rows.Scan(&schedule.GroupName, &groupId, &schedule.Title, &studentId, &trainerId, &schedule.Start, &schedule.End); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		if groupId.Valid {
			schedule.GroupId = groupId.UUID
		} else {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrInvalidUUID)
		}
		if studentId.Valid {
			schedule.StudentId = studentId.UUID
		} else {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrInvalidUUID)
		}
		if trainerId.Valid {
			schedule.TrainerId = trainerId.UUID
		} else {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrInvalidUUID)
		}

		schedules = append(schedules, &schedule)
	}

	logger.GetLoggerFromCtx(ctx).Debug(ctx, fmt.Sprintf("%d", len(schedules)))

	return schedules, nil
}

func (s *Storage) DeleteSchedule(ctx context.Context, groupId, trainerId uuid.UUID) error {
	const op = "psql.DeleteSchedule"

	query := `DELETE FROM schedules WHERE group_id = $1 AND trainer_id = $2`

	if _, err := s.db.Exec(ctx, query, groupId, trainerId); err != nil {
		// TODO: error

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
