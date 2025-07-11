package schedule

import (
	"context"

	"github.com/google/uuid"
	schedulev1 "github.com/hesoyamTM/apphelper-protos/gen/go/schedule"
	"github.com/hesoyamTM/apphelper-schedule/internal/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GroupService interface {
	AddToGroup(ctx context.Context, studentId uuid.UUID, link string) error
	CreateGroup(ctx context.Context, trainerId uuid.UUID, name string) error
	GetGroups(ctx context.Context, trainerId, studentId uuid.UUID) ([]*models.Group, error)
	DeleteGroup(ctx context.Context, groupId, trainerId uuid.UUID) error
}

func (s *serverAPI) AddToGroup(ctx context.Context, req *schedulev1.AddToGroupRequest) (*schedulev1.Empty, error) {
	userId, err := uuid.Parse(req.GetStudentId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "validation error")
	}
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
	userId, err := uuid.Parse(req.GetTrainerId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "validation error")
	}
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

func (s *serverAPI) GetGroups(ctx context.Context, req *schedulev1.GetGroupsRequest) (*schedulev1.GetGroupsResponse, error) {
	trainerId, err := uuid.Parse(req.GetTrainerId())
	if err != nil {
		trainerId = uuid.Nil
	}
	studentId, err := uuid.Parse(req.GetStudentId())
	if err != nil {
		studentId = uuid.Nil
	}

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
		students := make([]string, len(groups[i].Students))
		for j := range groups[i].Students {
			students[j] = groups[i].Students[j].String()
		}

		group := schedulev1.Group{
			Id:        groups[i].Id.String(),
			Name:      groups[i].Name,
			TrainerId: groups[i].TrainerId.String(),
			Students:  students,
			Link:      groups[i].Link,
		}

		groupsResp[i] = &group
	}

	return &schedulev1.GetGroupsResponse{
		Groups: groupsResp,
	}, nil
}

func (s *serverAPI) DeleteGroup(ctx context.Context, req *schedulev1.DeleteGroupRequest) (*schedulev1.Empty, error) {
	trainerId, err := uuid.Parse(req.GetTrainerId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "validation error")
	}
	groupId, err := uuid.Parse(req.GetGroupId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "validation error")
	}

	if err := s.group.DeleteGroup(ctx, groupId, trainerId); err != nil {
		// TODO: error

		return nil, status.Error(codes.Internal, "internal error")
	}
	return nil, nil
}
