package main

import (
	"context"
	"time"

	api "github.com/bloomingFlower/rssagg/protos"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/bloomingFlower/rssagg/internal/database"
	"github.com/google/uuid"
)

func (s *server) HandlerCreateFeed(ctx context.Context, req *api.CreateFeedRequest) (*api.Feed, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid user ID: %v", err)
	}

	feed, err := s.DB.CreateFeed(ctx, database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      req.Name,
		Url:       req.Url,
		UserID:    userID,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Couldn't create feed: %v", err)
	}

	lastFetchedAt := ""
	if feed.LastFetchedAt.Valid {
		lastFetchedAt = feed.LastFetchedAt.Time.Format(time.RFC3339)
	}

	return &api.Feed{
		Id:            feed.ID.String(),
		CreatedAt:     timestamppb.New(feed.CreatedAt),
		UpdatedAt:     timestamppb.New(feed.UpdatedAt),
		Name:          feed.Name,
		Url:           feed.Url,
		UserId:        feed.UserID.String(),
		LastFetchedAt: lastFetchedAt,
	}, nil
}

func (s *server) HandlerGetFeeds(req *api.GetFeedsRequest, stream api.ApiService_HandlerGetFeedsServer) error {
	feeds, err := s.DB.GetFeeds(context.Background())
	if err != nil {
		return status.Errorf(codes.Internal, "Couldn't get feeds: %v", err)
	}
	for _, feed := range feeds {
		err := stream.Send(&api.Feed{
			Id:            feed.ID.String(),
			CreatedAt:     timestamppb.New(feed.CreatedAt),
			UpdatedAt:     timestamppb.New(feed.UpdatedAt),
			Name:          feed.Name,
			Url:           feed.Url,
			UserId:        feed.UserID.String(),
			LastFetchedAt: feed.LastFetchedAt.Time.Format(time.RFC3339),
		})
		if err != nil {
			return status.Errorf(codes.Internal, "Error sending feed: %v", err)
		}
	}
	return nil
}
