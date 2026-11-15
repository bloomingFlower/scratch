package main

import (
	"context"
	"fmt"
	"time"

	api "github.com/bloomingFlower/rssagg/protos"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/bloomingFlower/rssagg/internal/database"
	"github.com/google/uuid"
)

func (s *server) HandlerCreateFeed(ctx context.Context, req *api.CreateFeedRequest) (*api.Feed, error) {
	userID, err := ownUserID(ctx, req.UserId)
	if err != nil {
		return nil, err
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
	ctx := stream.Context()
	userID, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	feeds, err := s.DB.GetFeeds(ctx)
	if err != nil {
		return status.Errorf(codes.Internal, "Couldn't get feeds: %v", err)
	}
	count := 0
	for _, feed := range feeds {
		if feed.UserID != userID {
			continue
		}
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
		count++
	}

	// 트레일러 설정
	trailer := metadata.Pairs(
		"total-feeds", fmt.Sprintf("%d", count),
		"status", "success",
	)
	stream.SetTrailer(trailer)

	return nil
}
