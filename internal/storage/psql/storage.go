package psql

import (
	"context"
	"fmt"
	"time"

	"github.com/hesoyamTM/apphelper-schedule/internal/models"
	"github.com/hesoyamTM/apphelper-sso/pkg/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

func New(host, user, password, db string, port int) *Storage {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", user, password, host, port, db)

	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		panic(err)
	}

	return &Storage{
		db: pool,
	}
}

func (s *Storage) CreateGroup(ctx context.Context, name, link string, trainerId int64) error {
	const op = "psql.CreateGroup"

	query := `INSERT INTO groups (name, trainer_id, student_ids, invitation_link)
	VALUES ($1, $2, array[]::int[], $3)`

	if _, err := s.db.Exec(ctx, query, name, trainerId, link); err != nil {
		// TODO: error
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) CreateSchedule(ctx context.Context, groupId, studentId, trainerId int64, date time.Time) error {
	const op = "psql.CreateSchedule"

	query := `INSERT INTO schedules (group_id, student_id, trainer_id, date)
	VALUES ($1, $2, $3, $4)`

	if _, err := s.db.Exec(ctx, query, groupId, studentId, trainerId, date); err != nil {
		// TODO: error

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) AddToGroup(ctx context.Context, studentId int64, link string) error {
	const op = "psql.AddToGroup"

	query := `
	UPDATE groups SET student_ids = array_append(student_ids, $1)
	WHERE invitation_link = $2
	AND NOT $1 = ANY(student_ids::int[])
	AND $1 != trainer_id`

	if _, err := s.db.Exec(ctx, query, studentId, link); err != nil {
		// TODO: error

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) ProvideGroups(ctx context.Context, trainerId, studentId int64) ([]models.Group, error) {
	if trainerId > 0 && studentId > 0 {
		return s.provideGroups(ctx, trainerId, studentId)
	}
	if trainerId > 0 {
		return s.provideGroupsByTrainer(ctx, trainerId)
	}
	if studentId > 0 {
		return s.provideGroupsByStudent(ctx, studentId)
	}

	return nil, nil
}

func (s *Storage) ProvideGroup(ctx context.Context, groupId int64) (models.Group, error) {
	//const op = "psql.provideGroup"

	query := `
	SELECT id, name, trainer_id, student_ids, invitation_link
	FROM groups
	WHERE id  = $1`

	row := s.db.QueryRow(ctx, query, groupId)

	var group models.Group
	row.Scan(&group.Id, &group.Name, &group.TrainerId, &group.Students, &group.Link)

	return group, nil
}

func (s *Storage) provideGroups(ctx context.Context, trainerId, studentId int64) ([]models.Group, error) {
	const op = "psql.provideGroups"

	query := `
	SELECT id, name, trainer_id, student_ids, invitation_link
	FROM groups
	WHERE trainer_id = $1 AND $2 = ANY(student_ids::int[])`

	rows, err := s.db.Query(ctx, query, trainerId, studentId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	groups := make([]models.Group, 0)
	for rows.Next() {
		var group models.Group
		if err := rows.Scan(&group.Id, &group.Name, &group.TrainerId, &group.Students, &group.Link); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		groups = append(groups, group)
	}

	return groups, nil
}

func (s *Storage) provideGroupsByTrainer(ctx context.Context, trainerId int64) ([]models.Group, error) {
	const op = "psql.provideGroupsByTrainer"

	query := `
	SELECT id, name, trainer_id, student_ids, invitation_link
	FROM groups
	WHERE trainer_id = $1`

	rows, err := s.db.Query(ctx, query, trainerId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	groups := make([]models.Group, 0)
	for rows.Next() {
		var group models.Group
		if err := rows.Scan(&group.Id, &group.Name, &group.TrainerId, &group.Students, &group.Link); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		groups = append(groups, group)
	}

	return groups, nil
}

func (s *Storage) provideGroupsByStudent(ctx context.Context, studentId int64) ([]models.Group, error) {
	const op = "psql.provideGroupsByStudent"

	query := `
	SELECT id, name, trainer_id, student_ids, invitation_link
	FROM groups
	WHERE $1 = ANY(student_ids::int[])`

	rows, err := s.db.Query(ctx, query, studentId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	groups := make([]models.Group, 0)
	for rows.Next() {
		var group models.Group
		if err := rows.Scan(&group.Id, &group.Name, &group.TrainerId, &group.Students, &group.Link); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		groups = append(groups, group)
	}

	return groups, nil
}

func (s *Storage) ProvideSchedules(ctx context.Context, trainerId, studentId int64) ([]models.Schedule, error) {
	if trainerId > 0 && studentId > 0 {
		query := `SELECT groups.name, schedules.group_id, schedules.student_id, schedules.trainer_id, schedules.date
		FROM schedules
		INNER JOIN groups ON groups.id = schedules.group_id
		WHERE schedules.trainer_id = $1 AND schedules.student_id = $2`
		return s.provideSchedules(ctx, func() (pgx.Rows, error) { return s.db.Query(ctx, query, trainerId, studentId) })
	}
	if trainerId > 0 {
		query := `SELECT groups.name, schedules.group_id, schedules.student_id, schedules.trainer_id, schedules.date
		FROM schedules
		INNER JOIN groups ON groups.id = schedules.group_id
		WHERE schedules.trainer_id = $1`
		return s.provideSchedules(ctx, func() (pgx.Rows, error) { return s.db.Query(ctx, query, trainerId) })
	}
	if studentId > 0 {
		query := `SELECT groups.name, schedules.group_id, schedules.student_id, schedules.trainer_id, schedules.date
		FROM schedules
		INNER JOIN groups ON groups.id = schedules.group_id
		WHERE schedules.student_id = $1`
		return s.provideSchedules(ctx, func() (pgx.Rows, error) { return s.db.Query(ctx, query, studentId) })
	}
	return nil, nil
}

func (s *Storage) provideSchedules(ctx context.Context, queryFunc func() (pgx.Rows, error)) ([]models.Schedule, error) {
	const op = "psql.provideSchedules"

	rows, err := queryFunc()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	schedules := make([]models.Schedule, 0)
	for rows.Next() {
		var schedule models.Schedule
		if err := rows.Scan(&schedule.GroupName, &schedule.GroupId, &schedule.StudentId, &schedule.TrainerId, &schedule.Date); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		schedules = append(schedules, schedule)
	}

	logger.GetLoggerFromCtx(ctx).Debug(ctx, fmt.Sprintf("%d", len(schedules)))

	return schedules, nil
}
