package main

import (
	"context"

	"github.com/bloomingFlower/rssagg/internal/auth"
	"github.com/bloomingFlower/rssagg/internal/database"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type userLookup func(context.Context, string) (database.User, error)

// Health probes are the only unary RPCs that do not require a credential.
func publicHealthMethod(method string) bool {
	return method == "/api.ApiService/handlerHealthz" || method == "/api.ApiService/CheckReadiness"
}

func authenticatedContext(ctx context.Context, lookup userLookup) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md["api_key"]) != 1 || md["api_key"][0] == "" {
		return nil, status.Error(codes.Unauthenticated, "API key required")
	}
	user, err := lookup(ctx, md["api_key"][0])
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid API key")
	}
	return auth.ContextWithUser(ctx, databaseUserToUser(user)), nil
}

func unaryAuth(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler, lookup userLookup) (interface{}, error) {
	if publicHealthMethod(info.FullMethod) {
		return handler(ctx, req)
	}
	ctx, err := authenticatedContext(ctx, lookup)
	if err != nil {
		return nil, err
	}
	return handler(ctx, req)
}

func (s *server) middlewareAuth(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return unaryAuth(ctx, req, info, handler, s.DB.GetUserByAPIKey)
}

type contextServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *contextServerStream) Context() context.Context { return s.ctx }

func streamAuth(srv interface{}, ss grpc.ServerStream, handler grpc.StreamHandler, lookup userLookup) error {
	ctx, err := authenticatedContext(ss.Context(), lookup)
	if err != nil {
		return err
	}
	return handler(srv, &contextServerStream{ServerStream: ss, ctx: ctx})
}

func (s *server) middlewareAuthStream(srv interface{}, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return streamAuth(srv, ss, handler, s.DB.GetUserByAPIKey)
}
