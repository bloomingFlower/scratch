package main

import (
	"context"
	api "github.com/bloomingFlower/rssagg/protos"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *server) Err(ctx context.Context, req *api.ErrRequest) (*api.ErrResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Err not implemented")
}
