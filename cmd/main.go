package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/hesoyamTM/apphelper-schedule/internal/app"
	"github.com/hesoyamTM/apphelper-schedule/internal/config"
	"github.com/hesoyamTM/apphelper-sso/pkg/logger"
)

func main() {
	ctx := context.Background()
	cfg := config.MustLoad()

	ctx, err := logger.New(ctx, cfg.Env)
	if err != nil {
		panic(err)
	}

	log := logger.GetLoggerFromCtx(ctx)
	log.Debug(ctx, "logger is working")

	application := app.New(ctx, cfg)
	go application.GrpcApp.MustRun(ctx)
	go application.Redpanda.Start(ctx)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	application.GrpcApp.Stop(ctx)
	application.Redpanda.Stop(ctx)
	log.Info(ctx, "application stopped")
}
