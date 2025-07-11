package schedule

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hesoyamTM/apphelper-schedule/internal/models"
	"github.com/hesoyamTM/apphelper-sso/pkg/logger"
	"go.uber.org/zap"
)

type ScheduleStorage interface {
	CreateSchedule(ctx context.Context, sched *models.Schedule) error
	ProvideSchedules(ctx context.Context, trainerId, studentId uuid.UUID) ([]*models.Schedule, error)
	DeleteSchedule(ctx context.Context, groupId, trainerId uuid.UUID) error
}

type GroupStorage interface {
	ProvideGroup(ctx context.Context, groupId uuid.UUID) (*models.Group, error)
}

type Redpanda interface {
	ScheduleCreatedEvent(ctx context.Context, schedule *models.Schedule) error
}

type Schedule struct {
	db  ScheduleStorage
	gDB GroupStorage

	calendarManager *CalendarManager
	redpanda        Redpanda
}

func New(ctx context.Context, db ScheduleStorage, gDB GroupStorage, calendarManager *CalendarManager, redpanda Redpanda) *Schedule {
	return &Schedule{
		db:              db,
		gDB:             gDB,
		calendarManager: calendarManager,
		redpanda:        redpanda,
	}
}

func (s *Schedule) CreateSchedule(ctx context.Context, sched *models.Schedule) error {
	const op = "schedule.CreateSchedule"
	log := logger.GetLoggerFromCtx(ctx)

	event := &models.CalendarEvent{
		Title: sched.Title,
		Start: sched.Start,
		End:   sched.End,
	}

	// begin TX
	if err := s.calendarManager.CreateEvent(ctx, sched.TrainerId, sched.GroupId, event); err != nil {
		log.Error(ctx, "failed to create event", zap.Error(err))

		// TODO: error

		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.calendarManager.CreateEvent(ctx, sched.StudentId, sched.GroupId, event); err != nil {
		log.Error(ctx, "failed to create event", zap.Error(err))

		// TODO: error

		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.db.CreateSchedule(ctx, sched); err != nil {
		log.Error(ctx, "failed to create schedule", zap.Error(err))

		// TODO: error

		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.redpanda.ScheduleCreatedEvent(ctx, sched); err != nil {
		log.Error(ctx, "failed to send schedule created event", zap.Error(err))

		// TODO: error

		return fmt.Errorf("%s: %w", op, err)
	}
	// end TX

	return nil
}

func (s *Schedule) CreateScheduleForGroup(ctx context.Context, groupId uuid.UUID, sched *models.Schedule) error {
	const op = "schedule.CreateScheduleForGroup"
	log := logger.GetLoggerFromCtx(ctx)

	group, err := s.gDB.ProvideGroup(ctx, groupId)
	if err != nil {
		log.Error(ctx, "failed to fetch group", zap.Error(err))

		// TODO: error

		return fmt.Errorf("%s: %w", op, err)
	}

	event := &models.CalendarEvent{
		Title: sched.Title,
		Start: sched.Start,
		End:   sched.End,
	}

	if err := s.calendarManager.CreateEvent(ctx, sched.TrainerId, groupId, event); err != nil {
		log.Error(ctx, "failed to create event", zap.Error(err))

		// TODO: error

		return fmt.Errorf("%s: %w", op, err)
	}

	for _, student := range group.Students {
		if err := s.calendarManager.CreateEvent(ctx, student, groupId, event); err != nil {
			log.Error(ctx, "failed to create event", zap.Error(err))

			// TODO: error

			return fmt.Errorf("%s: %w", op, err)
		}

		sched.StudentId = student

		if err := s.db.CreateSchedule(ctx, sched); err != nil {
			log.Error(ctx, "failed to create schedule", zap.Error(err))

			// TODO: error

			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}

func (s *Schedule) GetSchedules(ctx context.Context, groupId, trainerId, studentId uuid.UUID) ([]*models.Schedule, error) {
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

func (s *Schedule) DeleteSchedule(ctx context.Context, groupId, trainerId uuid.UUID) error {
	const op = "schedule.DeleteSchedule"
	log := logger.GetLoggerFromCtx(ctx)

	if err := s.db.DeleteSchedule(ctx, groupId, trainerId); err != nil {
		log.Error(ctx, "failed to delete schedule", zap.Error(err))

		// TODO: error

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Schedule) LoginURL(ctx context.Context, userID uuid.UUID) string {
	return s.calendarManager.LoginURL(ctx, userID)
}

func (s *Schedule) Authorize(ctx context.Context, userId uuid.UUID, authcode, state string) error {
	const op = "schedule.Authorize"
	log := logger.GetLoggerFromCtx(ctx)

	if err := s.calendarManager.Authorize(ctx, userId, authcode, state); err != nil {
		log.Error(ctx, "failed to authorize", zap.Error(err))

		// TODO: error

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Schedule) IsAuthorized(ctx context.Context, userId uuid.UUID) bool {
	return s.calendarManager.IsAuthorized(ctx, userId)
}
