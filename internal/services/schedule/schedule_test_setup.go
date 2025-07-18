package schedule

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hesoyamTM/apphelper-schedule/internal/models"
	"github.com/stretchr/testify/mock"
)

type MockCalendarManager struct {
	mock.Mock
}

func (m *MockCalendarManager) LoginURL(ctx context.Context, userID uuid.UUID) string {
	args := m.Called(ctx, userID)
	return args.String(0)
}

func (m *MockCalendarManager) Authorize(ctx context.Context, userId uuid.UUID, authcode, state string) error {
	args := m.Called(ctx, userId, authcode, state)
	return args.Error(0)
}

func (m *MockCalendarManager) IsAuthorized(ctx context.Context, userId uuid.UUID) bool {
	args := m.Called(ctx, userId)
	return args.Bool(0)
}

func (m *MockCalendarManager) CreateEvent(ctx context.Context, userId, groupId uuid.UUID, event *models.CalendarEvent) error {
	args := m.Called(ctx, userId, groupId, event)
	return args.Error(0)
}

func (m *MockCalendarManager) GetEvents(ctx context.Context, userId uuid.UUID, minTime, maxTime time.Time) (*[]*models.CalendarEvent, error) {
	args := m.Called(ctx, userId, minTime, maxTime)
	return args.Get(0).(*[]*models.CalendarEvent), args.Error(1)
}

func (m *MockCalendarManager) DeleteEvent(ctx context.Context, userId uuid.UUID, group_id uuid.UUID, eventId string) error {
	args := m.Called(ctx, userId, group_id, eventId)
	return args.Error(0)
}

type MockScheduleStorage struct {
	mock.Mock
}

func (m *MockScheduleStorage) CreateSchedule(ctx context.Context, sched *models.Schedule) error {
	args := m.Called(ctx, sched)
	return args.Error(0)
}

func (m *MockScheduleStorage) ProvideSchedules(ctx context.Context, trainerId, studentId uuid.UUID) ([]*models.Schedule, error) {
	args := m.Called(ctx, trainerId, studentId)
	return args.Get(0).([]*models.Schedule), args.Error(1)
}

func (m *MockScheduleStorage) DeleteSchedule(ctx context.Context, groupId, trainerId uuid.UUID) error {
	args := m.Called(ctx, groupId, trainerId)
	return args.Error(0)
}

type MockGroupStorage struct {
	mock.Mock
}

func (m *MockGroupStorage) ProvideGroup(ctx context.Context, groupId uuid.UUID) (*models.Group, error) {
	args := m.Called(ctx, groupId)
	return args.Get(0).(*models.Group), args.Error(1)
}

type MockRedpanda struct {
	mock.Mock
}

func (m *MockRedpanda) ScheduleCreatedEvent(ctx context.Context, schedule *models.Schedule) error {
	args := m.Called(ctx, schedule)
	return args.Error(0)
}
