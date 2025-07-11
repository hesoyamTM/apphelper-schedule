package groups

import (
	"context"
	"fmt"

	"github.com/brianvoe/gofakeit"
	"github.com/google/uuid"
	"github.com/hesoyamTM/apphelper-schedule/internal/clients/redpanda"
	"github.com/hesoyamTM/apphelper-schedule/internal/models"
	"github.com/hesoyamTM/apphelper-sso/pkg/logger"
	"go.uber.org/zap"
)

type GroupStorage interface {
	CreateGroup(ctx context.Context, name, link string, trainerId uuid.UUID) error
	AddToGroup(ctx context.Context, studentId uuid.UUID, link string) (*models.Group, error)
	ProvideGroups(ctx context.Context, trainerId, studentId uuid.UUID) ([]*models.Group, error)
	ProvideGroup(ctx context.Context, groupId uuid.UUID) (*models.Group, error)
	DeleteGroup(ctx context.Context, groupId uuid.UUID, trainerId uuid.UUID) error
}

type RedPanda interface {
	GroupAddedEvent(ctx context.Context, group *redpanda.GroupAddedEvent) error
}

type Groups struct {
	db       GroupStorage
	redpanda RedPanda
}

func New(ctx context.Context, db GroupStorage, redpanda RedPanda) *Groups {
	return &Groups{
		db:       db,
		redpanda: redpanda,
	}
}

func (g *Groups) AddToGroup(ctx context.Context, studentId uuid.UUID, link string) error {
	const op = "groups.AddToGroup"
	log := logger.GetLoggerFromCtx(ctx)

	group, err := g.db.AddToGroup(ctx, studentId, link)
	if err != nil {
		log.Error(ctx, "failed to create schedule", zap.Error(err))

		// TODO: error

		return fmt.Errorf("%s: %w", op, err)
	}

	if err := g.redpanda.GroupAddedEvent(ctx, &redpanda.GroupAddedEvent{
		GroupId:   group.Id.String(),
		GroupName: group.Name,
		TrainerId: group.TrainerId.String(),
		StudentId: studentId.String(),
		Link:      link,
	}); err != nil {
		log.Error(ctx, "failed to create schedule", zap.Error(err))

		// TODO: error

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (g *Groups) CreateGroup(ctx context.Context, trainerId uuid.UUID, name string) error {
	const op = "groups.CreateGroup"
	log := logger.GetLoggerFromCtx(ctx)

	link := gofakeit.Password(true, true, false, false, false, 20)

	if err := g.db.CreateGroup(ctx, name, link, trainerId); err != nil {
		log.Error(ctx, "failed to create schedule", zap.Error(err))

		// TODO: error

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (g *Groups) GetGroups(ctx context.Context, trainerId, studentId uuid.UUID) ([]*models.Group, error) {
	const op = "groups.GetGroups"
	log := logger.GetLoggerFromCtx(ctx)

	groups, err := g.db.ProvideGroups(ctx, trainerId, studentId)
	if err != nil {
		log.Error(ctx, "failed to create schedule", zap.Error(err))

		// TODO: error

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return groups, nil
}

func (g *Groups) DeleteGroup(ctx context.Context, groupId, trainerId uuid.UUID) error {
	const op = "groups.DeleteGroup"
	log := logger.GetLoggerFromCtx(ctx)

	if err := g.db.DeleteGroup(ctx, groupId, trainerId); err != nil {
		log.Error(ctx, "failed to create schedule", zap.Error(err))

		// TODO: error

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (g *Groups) GetGroup(ctx context.Context, groupId uuid.UUID) (*models.Group, error) {
	const op = "groups.ProvideGroup"
	log := logger.GetLoggerFromCtx(ctx)

	group, err := g.db.ProvideGroup(ctx, groupId)
	if err != nil {
		log.Error(ctx, "failed to create schedule", zap.Error(err))

		// TODO: error

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return group, nil
}
