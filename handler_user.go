package main

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/bloomingFlower/rssagg/internal/auth"
	api "github.com/bloomingFlower/rssagg/protos"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/bloomingFlower/rssagg/internal/database"
	"github.com/google/uuid"
)

func (s *server) HandlerCreateUser(ctx context.Context, req *api.CreateUserRequest) (*api.User, error) {
	user, err := s.DB.CreateUser(ctx, database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      req.Name,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Couldn't create user: %v", err)
	}

	return &api.User{
		Id:        user.ID.String(),
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
		Name:      user.Name,
	}, nil
}

//func (apiCfg *apiConfig) handlerGetUser(w http.ResponseWriter, user database.User) {
//	respondWithJSON(w, http.StatusOK, databaseUserToUser(user))
//}

func (s *server) HandlerGetUser(ctx context.Context, _ *api.GetUserRequest) (*api.User, error) {
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}
	return user, nil
}

//func (apiCfg *apiConfig) handlerGetPostsForUser(w http.ResponseWriter, r *http.Request, user database.User) {
//	posts, err := apiCfg.DB.GetPostsForUser(r.Context(), database.GetPostsForUserParams{
//		UserID: user.ID,
//		Limit:  10,
//	})
//	if err != nil {
//		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Couldn't find posts for user: %v", err))
//		return
//	}
//	respondWithJSON(w, http.StatusOK, databasePostsToPosts(posts))
//}

func (s *server) HandlerGetPostsForUser(req *api.GetPostsForUserRequest, stream api.ApiService_HandlerGetPostsForUserServer) error {
	// Limit parsing
	limit, err := strconv.Atoi(req.Limit)
	if err != nil || limit < 1 || limit > 100 {
		return status.Error(codes.InvalidArgument, "limit must be between 1 and 100")
	}

	userID, err := ownUserID(stream.Context(), req.UserId)
	if err != nil {
		return err
	}

	// Use stream context
	posts, err := s.DB.GetPostsForUser(stream.Context(), database.GetPostsForUserParams{
		UserID: userID,
		Limit:  int32(limit),
	})
	if err != nil {
		log.Printf("Error getting posts for user: %v", err)
		return status.Errorf(codes.Internal, "Couldn't get posts for user: %v", err)
	}
	if len(posts) == 0 {
		return status.Errorf(codes.NotFound, "No posts found for user")
	}

	// Send each post to the stream
	for _, post := range posts {
		log.Printf("Processing post: %+v", post)
		apiPost := databasePostToPost(post)
		if apiPost == nil {
			log.Printf("Failed to convert database post to API post")
			continue
		}
		if err := stream.Send(apiPost); err != nil {
			return status.Errorf(codes.Internal, "Failed to send post: %v", err)
		}
	}

	return nil
}
