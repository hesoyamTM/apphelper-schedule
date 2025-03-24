package groups

import (
	"context"
	"fmt"

	"github.com/brianvoe/gofakeit"
	"github.com/hesoyamTM/apphelper-schedule/internal/models"
	"github.com/hesoyamTM/apphelper-sso/pkg/logger"
	"go.uber.org/zap"
)

type GroupStorage interface {
	CreateGroup(ctx context.Context, name, link string, trainerId int64) error
	AddToGroup(ctx context.Context, studentId int64, link string) error
	ProvideGroups(ctx context.Context, trainerId, studentId int64) ([]models.Group, error)
}

type Groups struct {
	db GroupStorage
}

func New(ctx context.Context, db GroupStorage) *Groups {
	return &Groups{
		db: db,
	}
}

func (g *Groups) AddToGroup(ctx context.Context, studentId int64, link string) error {
	const op = "groups.AddToGroup"
	log := logger.GetLoggerFromCtx(ctx)

	if err := g.db.AddToGroup(ctx, studentId, link); err != nil {
		log.Error(ctx, "failed to create schedule", zap.Error(err))

		// TODO: error

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (g *Groups) CreateGroup(ctx context.Context, trainerId int64, name string) error {
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

func (g *Groups) GetGroups(ctx context.Context, trainerId, studentId int64) ([]models.Group, error) {
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
