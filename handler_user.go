package main

import (
	"context"
	api "github.com/bloomingFlower/rssagg/protos"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"

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

func (s *server) HandlerGetUser(ctx context.Context, req *api.GetUserRequest) (*api.User, error) {
	user, err := s.DB.GetUserByAPIKey(ctx, req.ApiKey)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Couldn't get user: %v", err)
	}

	return databaseUserToUser(user), nil
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
	posts, err := s.DB.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: uuid.MustParse(req.UserId),
		Limit:  10,
	})
	if err != nil {
		return status.Errorf(codes.Internal, "Couldn't get posts for user: %v", err)
	}
	for _, post := range posts {
		err := stream.Send(databasePostToPost(post))
		if err != nil {
			return err
		}
	}
	return nil
}
