package schedule

import (
	"context"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/hesoyamTM/apphelper-sso/pkg/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func CheckIdPermission(ctx context.Context, ids ...uuid.UUID) error {
	const op = "schedule.CheckIdPermission"
	log := logger.GetLoggerFromCtx(ctx)

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Error(ctx, fmt.Sprintf("%s: metadata not found", op))
		return status.Error(codes.Internal, "internal error")
	}

	mdUid, ok := md["uid"]
	if !ok {
		log.Error(ctx, fmt.Sprintf("%s: uid not found", op))
		return status.Error(codes.Unauthenticated, "internal error")
	}

	uid, err := uuid.Parse(mdUid[0])
	if err != nil {
		log.Error(ctx, fmt.Sprintf("%s: invalid uid: %s", op, err))
		return status.Error(codes.Internal, "internal error")
	}

	if len(ids) == 0 {
		log.Error(ctx, fmt.Sprintf("%s: ids is empty", op))
		return status.Error(codes.Internal, "internal error")
	}

	if slices.Contains(ids, uid) {
		return nil
	}

	log.Error(ctx, fmt.Sprintf("%s: permission denied", op))
	return status.Error(codes.PermissionDenied, "permission denied")
}
