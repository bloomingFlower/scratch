package main

import (
	"context"
	"log"

	"github.com/bloomingFlower/rssagg/internal/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

//type authedHander func(http.ResponseWriter, *http.Request, database.User)

//	func (apiCfg *apiConfig) middlewareAuth(handler authedHander) http.HandlerFunc {
//		return func(w http.ResponseWriter, r *http.Request) {
//			apiKey, err := auth.GetAPIKey(r.Header)
//			if err != nil {
//				respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Auth error: %v", err))
//				return
//			}
//
//			user, err := apiCfg.DB.GetUserByAPIKey(r.Context(), apiKey)
//			if err != nil {
//				respondWithError(w, http.StatusNotFound, fmt.Sprintf("Couldn't get user: %v", err))
//				return
//			}
//
//			handler(w, r, user)
//		}
//	}
var allowedMethods = map[string]bool{
	"/api.ApiService/handlerHealthz":          true,
	"/api.ApiService/handlerGetUser":          true,
	"/api.ApiService/handlerGetFeed":          true,
	"/api.ApiService/handlerGetPostsForUser":  true,
	"/api.ApiService/handlerGetFeedFollows":   true,
	"/api.ApiService/handlerCreateFeedFollow": true,
	"/api.ApiService/handlerCreateUser":       true,
}

func (s *server) middlewareAuth(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.Println("unary info.FullMethod:", info.FullMethod)
	if allowedMethods[info.FullMethod] {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	apiKeys, ok := md["api_key"]
	if !ok || len(apiKeys) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "api key is not provided")
	}

	apiKey := apiKeys[0]

	user, err := s.DB.GetUserByAPIKey(ctx, apiKey)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Couldn't get user: %v", err)
	}

	apiUser := databaseUserToUser(user)
	ctx = auth.ContextWithUser(ctx, apiUser)
	return handler(ctx, req)
}

// Ad-hoc helper struct to wrap a ServerStream with a new context
type contextServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// NewContextServerStream creates a new ServerStream with the given context
func newContextServerStream(stream grpc.ServerStream, ctx context.Context) *contextServerStream {
	return &contextServerStream{
		ServerStream: stream,
		ctx:          ctx,
	}
}

func (s *contextServerStream) Context() context.Context {
	return s.ctx
}

func (s *server) middlewareAuthStream(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	// Extract the context from the stream
	ctx := ss.Context()
	log.Println("stream info.FullMethod:", info.FullMethod)
	// Check if the method is one of the methods that should not pass through the interceptor
	if allowedMethods[info.FullMethod] {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return status.Errorf(codes.Unauthenticated, "metadata is not provided")
		}

		apiKeys, ok := md["api_key"]
		if !ok || len(apiKeys) == 0 {
			return status.Errorf(codes.Unauthenticated, "api key is not provided")
		}

		apiKey := apiKeys[0]

		user, err := s.DB.GetUserByAPIKey(ctx, apiKey)
		if err != nil {
			return status.Errorf(codes.NotFound, "Couldn't get user: %v", err)
		}

		apiUser := databaseUserToUser(user)
		ctx = auth.ContextWithUser(ctx, apiUser)
	}

	// Create a new context-aware ServerStream
	wrappedStream := grpc.ServerStream(newContextServerStream(ss, ctx))

	// Continue execution of the handler with the wrapped stream
	return handler(srv, wrappedStream)
}
