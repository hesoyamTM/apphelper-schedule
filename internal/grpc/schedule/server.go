package schedule

import (
	schedulev1 "github.com/hesoyamTM/apphelper-protos/gen/go/schedule"
	"google.golang.org/grpc"
)

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
