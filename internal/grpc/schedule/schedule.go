package schedule

import (
	"context"

	"github.com/google/uuid"
	schedulev1 "github.com/hesoyamTM/apphelper-protos/gen/go/schedule"
	"github.com/hesoyamTM/apphelper-schedule/internal/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ScheduleService interface {
	CreateSchedule(ctx context.Context, sched *models.Schedule) error
	CreateScheduleForGroup(ctx context.Context, groupId uuid.UUID, sched *models.Schedule) error
	GetSchedules(ctx context.Context, groupId, trainerId, studentId uuid.UUID) ([]*models.Schedule, error)
	DeleteSchedule(ctx context.Context, groupId, trainerId uuid.UUID) error
	LoginURL(ctx context.Context, userID uuid.UUID) string
	Authorize(ctx context.Context, userId uuid.UUID, authcode, state string) error
	IsAuthorized(ctx context.Context, userId uuid.UUID) bool
}

func (s *serverAPI) CreateSchedule(ctx context.Context, req *schedulev1.CreateScheduleRequest) (*schedulev1.Empty, error) {
	trainerId, err := uuid.Parse(req.GetTrainerId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "validation error")
	}
	studentId, err := uuid.Parse(req.GetStudentId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "validation error")
	}
	groupId, err := uuid.Parse(req.GetGroupId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "validation error")
	}
	if err := CheckIdPermission(ctx, trainerId); err != nil {
		return nil, err
	}

	sched := &models.Schedule{
		GroupId:   groupId,
		Title:     req.GetTitle(),
		TrainerId: trainerId,
		StudentId: studentId,
		Start:     req.Start.AsTime(),
		End:       req.End.AsTime(),
	}

	if err := s.schedule.CreateSchedule(ctx, sched); err != nil {
		// TODO: error

		return nil, status.Error(codes.Internal, "internal error")
	}
	return nil, nil
}

func (s *serverAPI) CreateScheduleForGroup(ctx context.Context, req *schedulev1.CreateScheduleForGroupRequest) (*schedulev1.Empty, error) {
	trainerId, err := uuid.Parse(req.GetTrainerId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "validation error")
	}
	groupId, err := uuid.Parse(req.GetGroupId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "validation error")
	}

	if err := CheckIdPermission(ctx, trainerId); err != nil {
		return nil, err
	}
	sched := &models.Schedule{
		GroupId:   groupId,
		Title:     req.GetTitle(),
		TrainerId: trainerId,
		Start:     req.Start.AsTime(),
		End:       req.End.AsTime(),
	}

	if err := s.schedule.CreateScheduleForGroup(ctx, groupId, sched); err != nil {
		// TODO: error

		return nil, status.Error(codes.Internal, "internal error")
	}
	return nil, nil
}

func (s *serverAPI) GetSchedule(ctx context.Context, req *schedulev1.GetSchedulesRequest) (*schedulev1.GetSchedulesResponse, error) {
	trainerId, err := uuid.Parse(req.GetTrainerId())
	if err != nil {
		trainerId = uuid.Nil
	}
	studentId, err := uuid.Parse(req.GetStudentId())
	if err != nil {
		studentId = uuid.Nil
	}
	groupId, err := uuid.Parse(req.GetGroupId())
	if err != nil {
		groupId = uuid.Nil
	}

	if err := CheckIdPermission(ctx, trainerId, studentId); err != nil {
		return nil, err
	}

	schedules, err := s.schedule.GetSchedules(ctx, groupId, trainerId, studentId)
	if err != nil {
		// TODO: error

		return nil, status.Error(codes.Internal, "internal error")
	}

	schedulesResp := make([]*schedulev1.Schedule, len(schedules))
	for i := range schedules {
		sched := schedulev1.Schedule{
			GroupId:   schedules[i].GroupId.String(),
			GroupName: schedules[i].GroupName,
			TrainerId: schedules[i].TrainerId.String(),
			StudentId: schedules[i].StudentId.String(),
			Start:     timestamppb.New(schedules[i].Start),
			End:       timestamppb.New(schedules[i].End),
		}

		schedulesResp[i] = &sched
	}

	return &schedulev1.GetSchedulesResponse{
		Schedules: schedulesResp,
	}, nil
}

func (s *serverAPI) DeleteSchedule(ctx context.Context, req *schedulev1.DeleteScheduleRequest) (*schedulev1.Empty, error) {
	trainerId, err := uuid.Parse(req.GetTrainerId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "validation error")
	}
	scheduleId, err := uuid.Parse(req.GetScheduleId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "validation error")
	}

	if err := s.schedule.DeleteSchedule(ctx, scheduleId, trainerId); err != nil {
		// TODO: error

		return nil, status.Error(codes.Internal, "internal error")
	}
	return nil, nil
}

func (s *serverAPI) GetLoginLink(ctx context.Context, req *schedulev1.Empty) (*schedulev1.GetLoginLinkResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
	}
	uid, ok := md["uid"]
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "uid is not provided")
	}
	userId, err := uuid.Parse(uid[0])
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "validation error")
	}

	return &schedulev1.GetLoginLinkResponse{
		LoginLink: s.schedule.LoginURL(ctx, userId),
	}, nil
}

func (s *serverAPI) LoginCallback(ctx context.Context, req *schedulev1.LoginCallbackRequest) (*schedulev1.Empty, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
	}
	uid, ok := md["uid"]
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "uid is not provided")
	}
	userId, err := uuid.Parse(uid[0])
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "validation error")
	}

	authcode := req.GetAuthCode()
	state := req.GetState()

	if err := s.schedule.Authorize(ctx, userId, authcode, state); err != nil {
		// TODO: error

		return nil, status.Error(codes.Internal, "internal error")
	}
	return nil, nil
}

func (s *serverAPI) IsAuthenticated(ctx context.Context, req *schedulev1.Empty) (*schedulev1.IsAuthenticatedResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
	}
	uid, ok := md["uid"]
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "uid is not provided")
	}
	userId, err := uuid.Parse(uid[0])
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "validation error")
	}

	if s.schedule.IsAuthorized(ctx, userId) {
		return &schedulev1.IsAuthenticatedResponse{
			IsAuthenticated: true,
		}, nil
	}
	return &schedulev1.IsAuthenticatedResponse{
		IsAuthenticated: false,
	}, nil
}
