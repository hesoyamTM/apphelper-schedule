package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hesoyamTM/apphelper-schedule/internal/storage"
	"github.com/redis/go-redis/v9"
)

func (s *Storage) SetState(ctx context.Context, userId uuid.UUID, state string, stateTTL time.Duration) error {
	const op = "redis.SetState"

	_, err := s.client.Set(ctx, userId.String(), state, stateTTL).Result()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetState(ctx context.Context, userId uuid.UUID) (string, error) {
	const op = "redis.GetState"

	res, err := s.client.Get(ctx, userId.String()).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("%s: %w", op, storage.ErrStateNotFound)
		}

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return res, nil
}

func (s *Storage) DeleteState(ctx context.Context, userId uuid.UUID) error {
	const op = "redis.DeleteState"

	_, err := s.client.Del(ctx, userId.String()).Result()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
