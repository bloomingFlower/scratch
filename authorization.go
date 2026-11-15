package main

import (
	"context"

	"github.com/bloomingFlower/rssagg/internal/auth"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func currentUserID(ctx context.Context) (uuid.UUID, error) {
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return uuid.Nil, status.Error(codes.Unauthenticated, "authentication required")
	}
	id, err := uuid.Parse(user.GetId())
	if err != nil {
		return uuid.Nil, status.Error(codes.Unauthenticated, "invalid authenticated user")
	}
	return id, nil
}

func ownUserID(ctx context.Context, requested string) (uuid.UUID, error) {
	current, err := currentUserID(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	id, err := uuid.Parse(requested)
	if err != nil {
		return uuid.Nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}
	if id != current {
		return uuid.Nil, status.Error(codes.PermissionDenied, "user does not own this resource")
	}
	return id, nil
}
