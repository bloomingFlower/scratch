package main

import (
	"context"
	"fmt"
	api "github.com/bloomingFlower/rssagg/protos"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"time"

	"github.com/bloomingFlower/rssagg/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (s *server) HandlerCreateFeedFollow(ctx context.Context, req *api.CreateFeedFollowRequest) (*api.FeedFollow, error) {
	feedFollow, err := s.DB.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    uuid.MustParse(req.UserId),
		FeedID:    uuid.MustParse(req.FeedId),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Couldn't create feed follow: %v", err)
	}

	return databaseFeedFollowToFeedFollow(feedFollow), nil
}

//func (apiCfg *apiConfig) handlerCreateFeedFollow(w http.ResponseWriter, r *http.Request, user database.User) {
//	type parameters struct {
//		FeedID uuid.UUID `json:"feed_id"`
//	}
//	decoder := json.NewDecoder(r.Body)
//	params := parameters{}
//	err := decoder.Decode(&params)
//	if err != nil {
//		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error parsing JSON: %v", err))
//		return
//	}
//	feedFollow, err := apiCfg.DB.CreateFeedFollow(r.Context(), database.CreateFeedFollowParams{
//		ID:        uuid.New(),
//		CreatedAt: time.Now().UTC(),
//		UpdatedAt: time.Now().UTC(),
//		UserID:    user.ID,
//		FeedID:    params.FeedID,
//	})
//	if err != nil {
//		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Couldn't create feed follow: %v", err))
//		return
//	}
//	respondWithJSON(w, http.StatusCreated, databaseFeedFollowToFeedFollow(feedFollow))
//}

func (s *server) HandlerGetFeedFollows(req *api.GetFeedFollowsRequest, stream api.ApiService_HandlerGetFeedFollowsServer) error {
	feedFollows, err := s.DB.GetFeedFollows(context.Background(), uuid.MustParse(req.UserId))
	if err != nil {
		return status.Errorf(codes.Internal, "Couldn't get feed follows: %v", err)
	}

	for _, feedFollow := range feedFollows {
		err := stream.Send(databaseFeedFollowToFeedFollow(feedFollow))
		if err != nil {
			return status.Errorf(codes.Internal, "Couldn't send feed follow: %v", err)
		}
	}

	return nil
}

//func (apiCfg *apiConfig) handlerGetFeedFollows(w http.ResponseWriter, r *http.Request, user database.User) {
//	feedFollows, err := apiCfg.DB.GetFeedFollows(r.Context(), user.ID)
//	if err != nil {
//		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Couldn't get feed follows: %v", err))
//		return
//	}
//	respondWithJSON(w, http.StatusCreated, databaseFeedFollowsToFeedFollows(feedFollows))
//
//}

func (apiCfg *apiConfig) handlerDeleteFeedFollows(w http.ResponseWriter, r *http.Request, user database.User) {
	feedFollowIDstr := chi.URLParam(r, "feedFollowID")
	feedFollowID, err := uuid.Parse(feedFollowIDstr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Couldn't parse feed follow id: %v", err))
		return
	}

	apiCfg.DB.DeleteFeedFollow(r.Context(), database.DeleteFeedFollowParams{
		ID:     feedFollowID,
		UserID: user.ID,
	})

	if err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Couldn't delete feed follow: %v", err))
		return
	}

	respondWithJSON(w, http.StatusOK, struct{}{})
}
