package app

import (
	"context"

	"github.com/hesoyamTM/apphelper-schedule/internal/app/grpcapp"
	"github.com/hesoyamTM/apphelper-schedule/internal/services/groups"
	"github.com/hesoyamTM/apphelper-schedule/internal/services/schedule"
	"github.com/hesoyamTM/apphelper-schedule/internal/storage/psql"
)

type GrpcOpts struct {
	Host string
	Port int
}

type PsqlOpts struct {
	Host     string
	Port     int
	User     string
	Password string
	DB       string
}

type App struct {
	GrpcApp grpcapp.App
}

func New(ctx context.Context, gOpts GrpcOpts, pOpts PsqlOpts) *App {
	db := psql.New(pOpts.Host, pOpts.User, pOpts.Password, pOpts.DB, pOpts.Port)

	scheduleService := schedule.New(ctx, db, db)
	groupService := groups.New(ctx, db)

	grpcApp := grpcapp.New(
		ctx,
		gOpts.Host,
		gOpts.Port,
		scheduleService,
		groupService,
	)

	return &App{
		GrpcApp: *grpcApp,
	}
}
