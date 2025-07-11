package psql

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/hesoyamTM/apphelper-schedule/internal/models"
	"github.com/hesoyamTM/apphelper-schedule/internal/storage"
	"github.com/jackc/pgx/v5"
)

func (s *Storage) CreateGroup(ctx context.Context, name, link string, trainerId uuid.UUID) error {
	const op = "psql.CreateGroup"

	query := `INSERT INTO groups (name, trainer_id, student_ids, invitation_link)
	VALUES ($1, $2, array[]::uuid[], $3) RETURNING id`

	if _, err := s.db.Exec(ctx, query, name, trainerId, link); err != nil {
		// TODO: error
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) AddToGroup(ctx context.Context, studentId uuid.UUID, link string) (*models.Group, error) {
	const op = "psql.AddToGroup"

	query := `
	UPDATE groups SET student_ids = array_append(student_ids, $1)
	WHERE invitation_link = $2
	AND NOT $1 = ANY(student_ids::int[])
	AND $1 != trainer_id
	RETURNING id, name, trainer_id, student_ids, invitation_link`

	row := s.db.QueryRow(ctx, query, studentId, link)

	var group models.Group
	var studentIds []uuid.NullUUID
	var trainerId uuid.NullUUID

	if err := row.Scan(&group.Id, &group.Name, &trainerId, &studentIds, &group.Link); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrGroupNotFound
		}
		return &group, fmt.Errorf("%s: %w", op, err)
	}

	group.Students = make([]uuid.UUID, len(studentIds))
	for i := range studentIds {
		if studentIds[i].Valid {
			group.Students[i] = studentIds[i].UUID
		}
	}
	if trainerId.Valid {
		group.TrainerId = trainerId.UUID
	} else {
		return &group, fmt.Errorf("%s: %w", op, storage.ErrInvalidUUID)
	}
	return &group, nil
}

func (s *Storage) ProvideGroups(ctx context.Context, trainerId, studentId uuid.UUID) ([]*models.Group, error) {
	if trainerId != uuid.Nil && studentId != uuid.Nil {
		return s.provideGroups(ctx, trainerId, studentId)
	}
	if trainerId != uuid.Nil {
		return s.provideGroupsByTrainer(ctx, trainerId)
	}
	if studentId != uuid.Nil {
		return s.provideGroupsByStudent(ctx, studentId)
	}

	return nil, nil
}

func (s *Storage) ProvideGroup(ctx context.Context, groupId uuid.UUID) (*models.Group, error) {
	const op = "psql.provideGroup"

	query := `
	SELECT id, name, trainer_id, student_ids, invitation_link
	FROM groups
	WHERE id  = $1`

	row := s.db.QueryRow(ctx, query, groupId)

	var group models.Group
	var studentIds []uuid.NullUUID
	var trainerId uuid.NullUUID

	if err := row.Scan(&group.Id, &group.Name, &trainerId, &studentIds, &group.Link); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrGroupNotFound
		}
		return &group, fmt.Errorf("%s: %w", op, err)
	}

	group.Students = make([]uuid.UUID, len(studentIds))
	for i := range studentIds {
		if studentIds[i].Valid {
			group.Students[i] = studentIds[i].UUID
		}
	}
	if trainerId.Valid {
		group.TrainerId = trainerId.UUID
	} else {
		return &group, fmt.Errorf("%s: %w", op, storage.ErrInvalidUUID)
	}

	return &group, nil
}

func (s *Storage) provideGroups(ctx context.Context, trainerId, studentId uuid.UUID) ([]*models.Group, error) {
	const op = "psql.provideGroups"

	query := `
	SELECT id, name, trainer_id, student_ids, invitation_link
	FROM groups
	WHERE trainer_id = $1 AND $2 = ANY(student_ids::int[])`

	rows, err := s.db.Query(ctx, query, trainerId, studentId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrGroupNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	groups := make([]*models.Group, 0)
	for rows.Next() {
		var group models.Group
		if err := rows.Scan(&group.Id, &group.Name, &group.TrainerId, &group.Students, &group.Link); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		groups = append(groups, &group)
	}

	return groups, nil
}

func (s *Storage) provideGroupsByTrainer(ctx context.Context, trainerId uuid.UUID) ([]*models.Group, error) {
	const op = "psql.provideGroupsByTrainer"

	query := `
	SELECT id, name, trainer_id, student_ids, invitation_link
	FROM groups
	WHERE trainer_id = $1`

	rows, err := s.db.Query(ctx, query, trainerId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrGroupNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	groups := make([]*models.Group, 0)
	for rows.Next() {
		var group models.Group
		if err := rows.Scan(&group.Id, &group.Name, &group.TrainerId, &group.Students, &group.Link); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		groups = append(groups, &group)
	}

	return groups, nil
}

func (s *Storage) provideGroupsByStudent(ctx context.Context, studentId uuid.UUID) ([]*models.Group, error) {
	const op = "psql.provideGroupsByStudent"

	query := `
	SELECT id, name, trainer_id, student_ids, invitation_link
	FROM groups
	WHERE $1 = ANY(student_ids::int[])`

	rows, err := s.db.Query(ctx, query, studentId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrGroupNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	groups := make([]*models.Group, 0)
	for rows.Next() {
		var group models.Group
		if err := rows.Scan(&group.Id, &group.Name, &group.TrainerId, &group.Students, &group.Link); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		groups = append(groups, &group)
	}

	return groups, nil
}

func (s *Storage) DeleteGroup(ctx context.Context, groupId uuid.UUID, trainerId uuid.UUID) error {
	const op = "psql.DeleteGroup"

	query := `DELETE FROM groups WHERE id = $1 AND trainer_id = $2`

	if _, err := s.db.Exec(ctx, query, groupId); err != nil {
		// TODO: error

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
