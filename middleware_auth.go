package main

import (
	"context"
	"github.com/bloomingFlower/rssagg/internal/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//type authedHander func(http.ResponseWriter, *http.Request, database.User)

//func (apiCfg *apiConfig) middlewareAuth(handler authedHander) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		apiKey, err := auth.GetAPIKey(r.Header)
//		if err != nil {
//			respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Auth error: %v", err))
//			return
//		}
//
//		user, err := apiCfg.DB.GetUserByAPIKey(r.Context(), apiKey)
//		if err != nil {
//			respondWithError(w, http.StatusNotFound, fmt.Sprintf("Couldn't get user: %v", err))
//			return
//		}
//
//		handler(w, r, user)
//	}
//}

func (s *server) middlewareAuth(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if info.FullMethod == "/ApiService/handlerHealthz" || info.FullMethod == "/ApiService/handlerGetUser" || info.FullMethod == "/ApiService/handlerGetFeed" || info.FullMethod == "/ApiService/handlerGetPostsForUser" || info.FullMethod == "/ApiService/handlerGetFeedFollows" || info.FullMethod == "/ApiService/handlerCreateFeedFollow" || info.FullMethod == "/ApiService/handlerCreateUser" {
		apiKey, err := auth.GetAPIKeyFromContext(ctx)
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
