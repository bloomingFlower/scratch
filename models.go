package main

import (
	"time"

	api "github.com/bloomingFlower/rssagg/protos"

	"github.com/bloomingFlower/rssagg/internal/database"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	APIKey    string    `json:"api_key"`
}

type Feed struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	Url       string    `json:"url"`
	UserID    uuid.UUID `json:"user_id"`
}

type FeedFollow struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    uuid.UUID `json:"user_id"`
	FeedID    uuid.UUID `json:"feed_id"`
}

type Post struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Title       string    `json:"title"`
	Description *string   `json:"description"`
	PublishedAt time.Time `json:"publishedAt"`
	Url         string    `json:"url"`
	FeedID      uuid.UUID `json:"feed_id"`
}

func databaseUserToUser(dbUser database.User) *api.User {
	createdAt := timestamppb.New(dbUser.CreatedAt)
	updatedAt := timestamppb.New(dbUser.UpdatedAt)
	return &api.User{
		Id:        dbUser.ID.String(),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Name:      dbUser.Name,
		ApiKey:    dbUser.ApiKey,
	}
}

func databaseFeedToFeed(dbFeed database.Feed) Feed {
	return Feed{
		ID:        dbFeed.ID,
		CreatedAt: dbFeed.CreatedAt,
		UpdatedAt: dbFeed.UpdatedAt,
		Name:      dbFeed.Name,
		Url:       dbFeed.Url,
		UserID:    dbFeed.UserID,
	}
}

func databaseFeedsToFeeds(dbFeeds []database.Feed) []Feed {
	feeds := []Feed{}
	for _, dbFeed := range dbFeeds {
		feeds = append(feeds, databaseFeedToFeed(dbFeed))
	}
	return feeds
}

func databaseFeedFollowToFeedFollow(dbFeedFollow database.FeedFollow) *api.FeedFollow {
	createdAt := timestamppb.New(dbFeedFollow.CreatedAt)
	updatedAt := timestamppb.New(dbFeedFollow.UpdatedAt)
	return &api.FeedFollow{
		Id:        dbFeedFollow.ID.String(),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		UserId:    dbFeedFollow.UserID.String(),
		FeedId:    dbFeedFollow.FeedID.String(),
	}
}

func databaseFeedFollowsToFeedFollows(feedFollows []database.FeedFollow) []*api.FeedFollow {
	result := make([]*api.FeedFollow, len(feedFollows))
	for i, feedFollow := range feedFollows {
		result[i] = databaseFeedFollowToFeedFollow(feedFollow)
	}
	return result
}

func databasePostToPost(dbPost database.Post) *api.Post {
	if dbPost == (database.Post{}) {
		return nil
	}
	var description *string
	if dbPost.Description.Valid {
		description = &dbPost.Description.String
	}
	createdAt := timestamppb.New(dbPost.CreatedAt)
	updatedAt := timestamppb.New(dbPost.UpdatedAt)
	publishedAt := timestamppb.New(dbPost.PublishedAt)

	return &api.Post{
		Id:          dbPost.ID.String(),
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		Title:       dbPost.Title,
		Description: *description,
		PublishedAt: publishedAt,
		Url:         dbPost.Url,
		FeedId:      dbPost.FeedID.String(),
	}
}

func databasePostsToPosts(dbPosts []database.Post) []*api.Post {
	posts := []*api.Post{}
	for _, dbPost := range dbPosts {
		posts = append(posts, databasePostToPost(dbPost))
	}
	return posts
}
