package app

import (
	"context"

	"github.com/hesoyamTM/apphelper-schedule/internal/app/grpcapp"
	"github.com/hesoyamTM/apphelper-schedule/internal/clients"
	"github.com/hesoyamTM/apphelper-schedule/internal/clients/redpanda"
	"github.com/hesoyamTM/apphelper-schedule/internal/config"
	"github.com/hesoyamTM/apphelper-schedule/internal/services/groups"
	"github.com/hesoyamTM/apphelper-schedule/internal/services/schedule"
	"github.com/hesoyamTM/apphelper-schedule/internal/storage/psql"
	"github.com/hesoyamTM/apphelper-schedule/internal/storage/redis"
)

type App struct {
	GrpcApp  grpcapp.App
	Redpanda *redpanda.RedPanda
}

func New(ctx context.Context, cfg *config.Config) *App {
	db := psql.New(cfg.Psql)
	if err := psql.RunMigrations(ctx, cfg.Psql); err != nil {
		panic(err)
	}

	redpanda, err := redpanda.NewRedPanda(ctx, cfg.Redpanda)
	if err != nil {
		panic(err)
	}

	groupService := groups.New(ctx, db, redpanda)

	calendarService := clients.New(ctx, cfg.GoogleCalendar)
	sessionStorage := redis.New(cfg.Redis)
	calendaerManager := schedule.NewCalendarManager(sessionStorage, sessionStorage, calendarService, db, groupService, cfg.StateTTL)

	scheduleService := schedule.New(ctx, db, db, calendaerManager, redpanda)

	grpcApp := grpcapp.New(
		ctx,
		cfg.Grpc.Host,
		cfg.Grpc.Port,
		scheduleService,
		groupService,
	)

	return &App{
		GrpcApp:  *grpcApp,
		Redpanda: redpanda,
	}
}
