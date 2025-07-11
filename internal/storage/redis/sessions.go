package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hesoyamTM/apphelper-schedule/internal/models"
	"github.com/hesoyamTM/apphelper-schedule/internal/storage"
	"github.com/redis/go-redis/v9"
)

func (s *Storage) SetSession(ctx context.Context, userId uuid.UUID, tok models.Token, ttl time.Duration) error {
	const op = "redis.SetSession"

	data := map[string]string{
		"access_token":  tok.AccessToken,
		"refresh_token": tok.RefreshToken,
	}
	if err := s.client.HSet(ctx, userId.String(), data).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) ProvideSession(ctx context.Context, userId uuid.UUID) (models.Token, error) {
	const op = "redis.ProvideSession"

	res, err := s.client.HGetAll(ctx, userId.String()).Result()
	if err != nil {
		if err == redis.Nil {
			return models.Token{}, fmt.Errorf("%s: %w", op, storage.ErrSessionNotFound)
		}

		return models.Token{}, fmt.Errorf("%s: %w", op, err)
	}

	return models.Token{
		AccessToken:  res["access_token"],
		RefreshToken: res["refresh_token"],
	}, nil
}

func (s *Storage) DeleteSession(ctx context.Context, userId uuid.UUID) error {
	const op = "redis.DeleteSession"

	_, err := s.client.Del(ctx, userId.String()).Result()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
