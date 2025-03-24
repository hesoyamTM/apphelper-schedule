package schedule

import (
	"context"
	"time"

	schedulev1 "github.com/hesoyamTM/apphelper-protos/gen/go/schedule"
	"github.com/hesoyamTM/apphelper-schedule/internal/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ScheduleService interface {
	CreateSchedule(ctx context.Context, groupId, trainerId, studentId int64, date time.Time) error
	CreateScheduleForGroup(ctx context.Context, groupId, trainerId int64, date time.Time) error
	GetSchedules(ctx context.Context, groupId, trainerId, studentId int64) ([]models.Schedule, error)
}

type GroupService interface {
	AddToGroup(ctx context.Context, studentId int64, link string) error
	CreateGroup(ctx context.Context, trainerId int64, name string) error
	GetGroups(ctx context.Context, trainerId, studentId int64) ([]models.Group, error)
}

type serverAPI struct {
	schedulev1.UnimplementedScheduleServer

	schedule ScheduleService
	group    GroupService
}

func RegisterServer(grpcServer *grpc.Server, schedule ScheduleService, group GroupService) {
	schedulev1.RegisterScheduleServer(
		grpcServer,
		&serverAPI{
			schedule: schedule,
			group:    group,
		})
}

func (s *serverAPI) AddToGroup(ctx context.Context, req *schedulev1.AddToGroupRequest) (*schedulev1.Empty, error) {
	userId := req.GetStudentId()
	link := req.GetLink()

	if link == "" {
		return nil, status.Error(codes.InvalidArgument, "validation error")
	}

	if err := CheckIdPermission(ctx, userId); err != nil {
		return nil, err
	}

	if err := s.group.AddToGroup(ctx, userId, link); err != nil {
		// TODO: error

		return nil, status.Error(codes.Internal, "internal error")
	}
	return nil, nil
}

func (s *serverAPI) CreateGroup(ctx context.Context, req *schedulev1.CreateGroupRequest) (*schedulev1.Empty, error) {
	userId := req.GetTrainerId()
	name := req.GetName()

	if err := CheckIdPermission(ctx, userId); err != nil {
		return nil, err
	}

	if err := s.group.CreateGroup(ctx, userId, name); err != nil {
		// TODO: error

		return nil, status.Error(codes.Internal, "internal error")
	}
	return nil, nil
}

func (s *serverAPI) CreateSchedule(ctx context.Context, req *schedulev1.CreateScheduleRequest) (*schedulev1.Empty, error) {
	trainerId := req.GetTrainerId()
	studentId := req.GetStudentId()
	groupId := req.GetGroupId()
	date := req.GetDate().AsTime()

	if err := CheckIdPermission(ctx, trainerId); err != nil {
		return nil, err
	}

	if err := s.schedule.CreateSchedule(ctx, groupId, trainerId, studentId, date); err != nil {
		// TODO: error

		return nil, status.Error(codes.Internal, "internal error")
	}
	return nil, nil
}

func (s *serverAPI) CreateScheduleForGroup(ctx context.Context, req *schedulev1.CreateScheduleForGroupRequest) (*schedulev1.Empty, error) {
	trainerId := req.GetTrainerId()
	groupId := req.GetGroupId()
	date := req.GetDate().AsTime()

	if err := CheckIdPermission(ctx, trainerId); err != nil {
		return nil, err
	}

	if err := s.schedule.CreateScheduleForGroup(ctx, groupId, trainerId, date); err != nil {
		// TODO: error

		return nil, status.Error(codes.Internal, "internal error")
	}
	return nil, nil
}

func (s *serverAPI) GetSchedule(ctx context.Context, req *schedulev1.GetSchedulesRequest) (*schedulev1.GetSchedulesResponse, error) {
	trainerId := req.GetTrainerId()
	studentId := req.GetStudentId()
	groupId := req.GetGroupId()

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
			GroupId:   schedules[i].GroupId,
			GroupName: schedules[i].GroupName,
			TrainerId: schedules[i].TrainerId,
			StudentId: schedules[i].StudentId,
			Date:      timestamppb.New(schedules[i].Date),
		}

		schedulesResp[i] = &sched
	}

	return &schedulev1.GetSchedulesResponse{
		Schedules: schedulesResp,
	}, nil
}

func (s *serverAPI) GetGroups(ctx context.Context, req *schedulev1.GetGroupsRequest) (*schedulev1.GetGroupsResponse, error) {
	trainerId := req.GetTrainerId()
	studentId := req.GetStudentId()

	if err := CheckIdPermission(ctx, trainerId, studentId); err != nil {
		return nil, err
	}

	groups, err := s.group.GetGroups(ctx, trainerId, studentId)
	if err != nil {
		// TODO: error

		return nil, status.Error(codes.Internal, "internal error")
	}

	groupsResp := make([]*schedulev1.Group, len(groups))
	for i := range groups {
		group := schedulev1.Group{
			Id:        groups[i].Id,
			Name:      groups[i].Name,
			TrainerId: groups[i].TrainerId,
			Students:  groups[i].Students,
			Link:      groups[i].Link,
		}

		groupsResp[i] = &group
	}

	return &schedulev1.GetGroupsResponse{
		Groups: groupsResp,
	}, nil
}
