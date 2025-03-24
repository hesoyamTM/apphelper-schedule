package grpcapp

import (
	"context"
	"fmt"
	"net"

	"github.com/hesoyamTM/apphelper-schedule/internal/grpc/schedule"
	"github.com/hesoyamTM/apphelper-sso/pkg/logger"
	"google.golang.org/grpc"
)

type App struct {
	grpcServer *grpc.Server

	host string
	port int
}

func New(
	ctx context.Context,
	host string,
	port int,
	shedServ schedule.ScheduleService,
	groupServ schedule.GroupService) *App {

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(logger.LoggingInterceptor(ctx)))

	schedule.RegisterServer(grpcServer, shedServ, groupServ)

	return &App{
		host:       host,
		port:       port,
		grpcServer: grpcServer,
	}
}

func (a *App) MustRun(ctx context.Context) {
	log := logger.GetLoggerFromCtx(ctx)

	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", a.host, a.port))
	if err != nil {
		panic(err)
	}

	log.Info(ctx, "server is running")

	if err := a.grpcServer.Serve(l); err != nil {
		panic(err)
	}
}

func (a *App) Stop(ctx context.Context) {
	log := logger.GetLoggerFromCtx(ctx)

	log.Info(ctx, "grpc server is stopping")

	a.grpcServer.GracefulStop()
}
