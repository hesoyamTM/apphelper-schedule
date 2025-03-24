package schedule

import (
	"context"
	"fmt"
	"time"

	"github.com/hesoyamTM/apphelper-schedule/internal/models"
	"github.com/hesoyamTM/apphelper-sso/pkg/logger"
	"go.uber.org/zap"
)

type ScheduleStorage interface {
	CreateSchedule(ctx context.Context, groupId, studentId, trainerId int64, date time.Time) error
	ProvideSchedules(ctx context.Context, trainerId, studentId int64) ([]models.Schedule, error)
}

type GroupStorage interface {
	ProvideGroup(ctx context.Context, groupId int64) (models.Group, error)
}

type Schedule struct {
	db  ScheduleStorage
	gDB GroupStorage
}

func New(ctx context.Context, db ScheduleStorage, gDB GroupStorage) *Schedule {
	return &Schedule{
		db:  db,
		gDB: gDB,
	}
}

func (s *Schedule) CreateSchedule(ctx context.Context, groupId, trainerId, studentId int64, date time.Time) error {
	const op = "schedule.CreateSchedule"
	log := logger.GetLoggerFromCtx(ctx)

	if err := s.db.CreateSchedule(ctx, groupId, studentId, trainerId, date); err != nil {
		log.Error(ctx, "failed to create schedule", zap.Error(err))

		// TODO: error

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Schedule) CreateScheduleForGroup(ctx context.Context, groupId, trainerId int64, date time.Time) error {
	const op = "schedule.CreateScheduleForGroup"
	log := logger.GetLoggerFromCtx(ctx)

	group, err := s.gDB.ProvideGroup(ctx, groupId)
	if err != nil {
		log.Error(ctx, "failed to fetch group", zap.Error(err))

		// TODO: error

		return fmt.Errorf("%s: %w", op, err)
	}

	for _, student := range group.Students {
		if err := s.db.CreateSchedule(ctx, groupId, student, trainerId, date); err != nil {
			log.Error(ctx, "failed to create schedule", zap.Error(err))

			// TODO: error

			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}

func (s *Schedule) GetSchedules(ctx context.Context, groupId, trainerId, studentId int64) ([]models.Schedule, error) {
	const op = "schedule.GetSchedules"
	log := logger.GetLoggerFromCtx(ctx)

	schedules, err := s.db.ProvideSchedules(ctx, trainerId, studentId)
	if err != nil {
		log.Error(ctx, "failed to get schedule", zap.Error(err))

		// TODO: error

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return schedules, nil
}
