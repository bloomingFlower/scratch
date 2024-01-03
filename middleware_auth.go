package main

import (
	"context"
	"github.com/bloomingFlower/rssagg/internal/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
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
	"/ApiService/handlerHealthz":          true,
	"/ApiService/handlerGetUser":          true,
	"/ApiService/handlerGetFeed":          true,
	"/ApiService/handlerGetPostsForUser":  true,
	"/ApiService/handlerGetFeedFollows":   true,
	"/ApiService/handlerCreateFeedFollow": true,
	"/ApiService/handlerCreateUser":       true,
}

func (s *server) middlewareAuth(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.Println("unary info.FullMethod:", info.FullMethod)
	if allowedMethods[info.FullMethod] {
		return handler(ctx, req)
	} else {
		apiKey, err := auth.GetAPIKeyFromContext(ctx)
		log.Printf("API key: %s", apiKey)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "Auth error: %v", err)
		}

		user, err := s.DB.GetUserByAPIKey(ctx, apiKey)
		if err != nil {
			return nil, status.Errorf(codes.NotFound, "Couldn't get user: %v", err)
		}

		apiUser := databaseUserToUser(user)
		ctx = auth.ContextWithUser(ctx, apiUser)
	}
	return handler(ctx, req)
}

func (s *server) middlewareAuthStream(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	// Extract the context from the stream
	ctx := ss.Context()
	log.Println("stream info.FullMethod:", info.FullMethod)
	// Check if the method is one of the methods that should not pass through the interceptor
	if allowedMethods[info.FullMethod] {
		apiKey, err := auth.GetAPIKeyFromContext(ctx)
		if err != nil {
			return status.Errorf(codes.Unauthenticated, "Auth error: %v", err)
		}

		user, err := s.DB.GetUserByAPIKey(ctx, apiKey)
		if err != nil {
			return status.Errorf(codes.NotFound, "Couldn't get user: %v", err)
		}

		apiUser := databaseUserToUser(user)
		ctx = auth.ContextWithUser(ctx, apiUser)
	}

	// Continue execution of the handler
	return handler(srv, ss)
}
