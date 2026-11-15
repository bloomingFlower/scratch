package main

import (
	"context"
	"io"
	"net"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bloomingFlower/rssagg/internal/database"
	api "github.com/bloomingFlower/rssagg/protos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func TestGRPCAuthenticationAndBookmarkOwnership(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := &server{DB: database.New(db)}
	listener := bufconn.Listen(1 << 20)
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(s.middlewareAuth), grpc.StreamInterceptor(s.middlewareAuthStream))
	api.RegisterApiServiceServer(grpcServer, s)
	go func() { _ = grpcServer.Serve(listener) }()
	t.Cleanup(grpcServer.Stop)
	conn, err := grpc.NewClient("passthrough:///in-memory", grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	client := api.NewApiServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	feeds, err := client.HandlerGetFeeds(ctx, &api.GetFeedsRequest{})
	if err == nil {
		_, err = feeds.Recv()
	}
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("unauthenticated feed stream: %v", err)
	}
	if _, err := client.HandlerCreateFeedFollow(ctx, &api.CreateFeedFollowRequest{}); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("unauthenticated write: %v", err)
	}

	now := time.Now().UTC()
	authQuery := regexp.QuoteMeta("SELECT id, created_at, updated_at, name, api_key FROM users WHERE api_key = $1")
	expectAuth := func() {
		mock.ExpectQuery(authQuery).WithArgs("valid-test-key").WillReturnRows(
			sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "api_key"}).
				AddRow(ownerID, now, now, "owner", "valid-test-key"))
	}
	ownerCtx := metadata.AppendToOutgoingContext(ctx, "api_key", "valid-test-key")
	expectAuth()
	posts, err := client.HandlerGetPostsForUser(ownerCtx, &api.GetPostsForUserRequest{UserId: otherID, Limit: "12"})
	if err == nil {
		_, err = posts.Recv()
	}
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("another user's bookmarks: %v", err)
	}

	expectAuth()
	mock.ExpectQuery("SELECT posts.id, posts.created_at").WithArgs(ownerID, int64(12)).WillReturnRows(
		sqlmock.NewRows([]string{"id", "created_at", "updated_at", "title", "description", "published_at", "url", "feed_id"}).
			AddRow(otherID, now, now, "public bookmark", "description", now, "https://example.com", otherID))
	posts, err = client.HandlerGetPostsForUser(ownerCtx, &api.GetPostsForUserRequest{UserId: ownerID, Limit: "12"})
	if err != nil {
		t.Fatal(err)
	}
	post, err := posts.Recv()
	if err != nil || post.GetTitle() != "public bookmark" {
		t.Fatalf("owner bookmark: %v", err)
	}
	if _, err := posts.Recv(); err != io.EOF {
		t.Fatalf("expected stream completion: %v", err)
	}

	expectAuth()
	user, err := client.HandlerGetUser(ownerCtx, &api.GetUserRequest{})
	if err != nil || user.GetId() != ownerID || user.GetApiKey() != "" {
		t.Fatalf("user response leaked a credential or returned the wrong user: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
