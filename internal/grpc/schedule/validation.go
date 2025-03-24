package schedule

import (
	"context"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func CheckIdPermission(ctx context.Context, ids ...int64) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Internal, "internal error")
	}

	mdUid, ok := md["uid"]
	if !ok {
		return status.Error(codes.Unauthenticated, "internal error")
	}

	uid, err := strconv.ParseInt(mdUid[0], 10, 64)
	if err != nil {
		return status.Error(codes.Internal, "internal error")
	}

	if len(ids) == 0 {
		return status.Error(codes.Internal, "internal error")
	}

	for _, id := range ids {
		if id == uid {
			return nil
		}
	}

	return status.Error(codes.PermissionDenied, "permission denied")
}
