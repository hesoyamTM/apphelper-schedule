package schedule

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hesoyamTM/apphelper-schedule/internal/models"
	"github.com/hesoyamTM/apphelper-sso/pkg/logger"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateSchedule(t *testing.T) {
	MockCalendarManager := &MockCalendarManager{}
	MockScheduleStorage := &MockScheduleStorage{}
	MockGroupStorage := &MockGroupStorage{}
	MockRedpanda := &MockRedpanda{}

	MockCalendarManager.On("CreateEvent", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	MockScheduleStorage.On("CreateSchedule", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	MockRedpanda.On("ScheduleCreatedEvent", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	ctx, err := logger.New(context.Background(), "dev")
	if err != nil {
		t.Errorf("logger.New() error = %v", err)
	}

	s := New(ctx, MockScheduleStorage, MockGroupStorage, MockCalendarManager, MockRedpanda)

	trainerId := uuid.New()
	studentId := uuid.New()
	groupId := uuid.New()

	sched := &models.Schedule{
		Title:     "test",
		Start:     time.Now(),
		End:       time.Now().Add(time.Hour),
		TrainerId: trainerId,
		StudentId: studentId,
		GroupId:   groupId,
	}

	if err := s.CreateSchedule(ctx, sched); err != nil {
		t.Errorf("CreateSchedule() error = %v", err)
	}

	MockCalendarManager.AssertCalled(t, "CreateEvent", ctx, trainerId, groupId, mock.Anything)
	MockCalendarManager.AssertCalled(t, "CreateEvent", ctx, studentId, groupId, mock.Anything)
	MockScheduleStorage.AssertCalled(t, "CreateSchedule", ctx, sched)
	MockRedpanda.AssertCalled(t, "ScheduleCreatedEvent", ctx, sched)
}

func TestCreateScheduleForGroup(t *testing.T) {
	MockCalendarManager := &MockCalendarManager{}
	MockScheduleStorage := &MockScheduleStorage{}
	MockGroupStorage := &MockGroupStorage{}
	MockRedpanda := &MockRedpanda{}

	studentId := uuid.New()

	MockCalendarManager.On("CreateEvent", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	MockScheduleStorage.On("CreateSchedule", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	MockRedpanda.On("ScheduleCreatedEvent", mock.Anything, mock.Anything).Return(nil)
	MockGroupStorage.On("ProvideGroup", mock.Anything, mock.Anything).Return(&models.Group{
		Students: []uuid.UUID{studentId},
	}, nil)

	ctx, err := logger.New(context.Background(), "dev")
	if err != nil {
		t.Errorf("logger.New() error = %v", err)
	}

	s := New(ctx, MockScheduleStorage, MockGroupStorage, MockCalendarManager, MockRedpanda)

	groupId := uuid.New()
	trainerId := uuid.New()

	sched := &models.Schedule{
		Title:     "test",
		Start:     time.Now(),
		End:       time.Now().Add(time.Hour),
		TrainerId: trainerId,
		GroupId:   groupId,
	}

	if err := s.CreateScheduleForGroup(ctx, groupId, sched); err != nil {
		t.Errorf("CreateScheduleForGroup() error = %v", err)
	}

	MockCalendarManager.AssertCalled(t, "CreateEvent", ctx, trainerId, groupId, mock.Anything)
	MockCalendarManager.AssertCalled(t, "CreateEvent", ctx, studentId, groupId, mock.Anything)
	MockScheduleStorage.AssertCalled(t, "CreateSchedule", ctx, sched)
	MockRedpanda.AssertCalled(t, "ScheduleCreatedEvent", ctx, sched)
}

func TestGetScheduleByTrainer(t *testing.T) {
	MockCalendarManager := &MockCalendarManager{}
	MockScheduleStorage := &MockScheduleStorage{}
	MockGroupStorage := &MockGroupStorage{}
	MockRedpanda := &MockRedpanda{}

	groupId := uuid.New()
	trainerId := uuid.New()

	MockScheduleStorage.On("ProvideSchedules", mock.Anything, mock.Anything, mock.Anything).Return([]*models.Schedule{
		{
			Title:     "test",
			Start:     time.Now(),
			End:       time.Now().Add(time.Hour),
			TrainerId: trainerId,
			StudentId: uuid.New(),
			GroupId:   groupId,
		},
	}, nil)

	ctx, err := logger.New(context.Background(), "dev")
	if err != nil {
		t.Errorf("logger.New() error = %v", err)
	}

	s := New(ctx, MockScheduleStorage, MockGroupStorage, MockCalendarManager, MockRedpanda)

	schedules, err := s.GetSchedules(ctx, groupId, trainerId, uuid.Nil)
	if err != nil {
		t.Errorf("GetSchedules() error = %v", err)
	}

	require.Equal(t, 1, len(schedules))
	require.Equal(t, "test", schedules[0].Title)
	require.Equal(t, trainerId, schedules[0].TrainerId)

	MockScheduleStorage.AssertCalled(t, "ProvideSchedules", ctx, trainerId, mock.Anything)
}

func TestGetScheduleByStudent(t *testing.T) {
	MockCalendarManager := &MockCalendarManager{}
	MockScheduleStorage := &MockScheduleStorage{}
	MockGroupStorage := &MockGroupStorage{}
	MockRedpanda := &MockRedpanda{}

	groupId := uuid.New()
	StudentId := uuid.New()

	MockScheduleStorage.On("ProvideSchedules", mock.Anything, mock.Anything, mock.Anything).Return([]*models.Schedule{
		{
			Title:     "test",
			Start:     time.Now(),
			End:       time.Now().Add(time.Hour),
			TrainerId: uuid.New(),
			StudentId: StudentId,
			GroupId:   groupId,
		},
	}, nil)

	ctx, err := logger.New(context.Background(), "dev")
	if err != nil {
		t.Errorf("logger.New() error = %v", err)
	}

	s := New(ctx, MockScheduleStorage, MockGroupStorage, MockCalendarManager, MockRedpanda)

	schedules, err := s.GetSchedules(ctx, groupId, uuid.Nil, StudentId)
	if err != nil {
		t.Errorf("GetSchedules() error = %v", err)
	}

	require.Equal(t, 1, len(schedules))
	require.Equal(t, "test", schedules[0].Title)
	require.Equal(t, StudentId, schedules[0].StudentId)

	MockScheduleStorage.AssertCalled(t, "ProvideSchedules", ctx, mock.Anything, StudentId)
}

func TestGetScheduleByTrainerAndStudent(t *testing.T) {
	MockCalendarManager := &MockCalendarManager{}
	MockScheduleStorage := &MockScheduleStorage{}
	MockGroupStorage := &MockGroupStorage{}
	MockRedpanda := &MockRedpanda{}

	groupId := uuid.New()
	StudentId := uuid.New()
	TrainerId := uuid.New()

	MockScheduleStorage.On("ProvideSchedules", mock.Anything, mock.Anything, mock.Anything).Return([]*models.Schedule{
		{
			Title:     "test",
			Start:     time.Now(),
			End:       time.Now().Add(time.Hour),
			TrainerId: TrainerId,
			StudentId: StudentId,
			GroupId:   groupId,
		},
	}, nil)

	ctx, err := logger.New(context.Background(), "dev")
	if err != nil {
		t.Errorf("logger.New() error = %v", err)
	}

	s := New(ctx, MockScheduleStorage, MockGroupStorage, MockCalendarManager, MockRedpanda)

	schedules, err := s.GetSchedules(ctx, groupId, TrainerId, StudentId)
	if err != nil {
		t.Errorf("GetSchedules() error = %v", err)
	}

	require.Equal(t, 1, len(schedules))
	require.Equal(t, "test", schedules[0].Title)
	require.Equal(t, StudentId, schedules[0].StudentId)
	require.Equal(t, TrainerId, schedules[0].TrainerId)

	MockScheduleStorage.AssertCalled(t, "ProvideSchedules", ctx, TrainerId, StudentId)
}

func TestDeleteSchedule(t *testing.T) {
	MockCalendarManager := &MockCalendarManager{}
	MockScheduleStorage := &MockScheduleStorage{}
	MockGroupStorage := &MockGroupStorage{}
	MockRedpanda := &MockRedpanda{}

	groupId := uuid.New()
	trainerId := uuid.New()

	MockScheduleStorage.On("DeleteSchedule", mock.Anything, groupId, trainerId).Return(nil)

	ctx, err := logger.New(context.Background(), "dev")
	if err != nil {
		t.Errorf("logger.New() error = %v", err)
	}

	s := New(ctx, MockScheduleStorage, MockGroupStorage, MockCalendarManager, MockRedpanda)

	if err := s.DeleteSchedule(ctx, groupId, trainerId); err != nil {
		t.Errorf("DeleteSchedule() error = %v", err)
	}

	MockScheduleStorage.AssertCalled(t, "DeleteSchedule", ctx, groupId, trainerId)
}
