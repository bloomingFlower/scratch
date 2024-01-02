package main

import (
	"context"
	api "github.com/bloomingFlower/rssagg/protos"
)

func (s *server) HandlerReadiness(ctx context.Context, req *api.ReadinessRequest) (*api.ReadinessResponse, error) {
	return &api.ReadinessResponse{
		IsReady: true,
	}, nil
}
