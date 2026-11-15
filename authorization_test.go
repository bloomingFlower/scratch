package main

import (
	"context"
	"errors"
	"testing"

	"github.com/bloomingFlower/rssagg/internal/auth"
	"github.com/bloomingFlower/rssagg/internal/database"
	api "github.com/bloomingFlower/rssagg/protos"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const ownerID = "ba1af24d-9bfc-4f40-8c9c-9c1ea87b69fa"
const otherID = "00000000-0000-4000-8000-000000000001"

func testLookup(_ context.Context, key string) (database.User, error) {
	if key != "valid-test-key" {
		return database.User{}, errors.New("not found")
	}
	return database.User{ID: uuid.MustParse(ownerID), Name: "owner", ApiKey: key}, nil
}

type testStream struct{ ctx context.Context }

func (s *testStream) Context() context.Context     { return s.ctx }
func (s *testStream) SetHeader(metadata.MD) error  { return nil }
func (s *testStream) SendHeader(metadata.MD) error { return nil }
func (s *testStream) SetTrailer(metadata.MD)       {}
func (s *testStream) SendMsg(interface{}) error    { return nil }
func (s *testStream) RecvMsg(interface{}) error    { return nil }
func (s *testStream) Send(*api.Post) error         { return nil }

func TestUnaryAndStreamAuthentication(t *testing.T) {
	var seen bool
	handler := func(ctx context.Context, _ interface{}) (interface{}, error) {
		seen = true
		user, ok := auth.UserFromContext(ctx)
		if !ok || user.GetId() != ownerID || user.GetApiKey() != "" {
			t.Fatal("handler did not receive a key-free authenticated identity")
		}
		return nil, nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/api.ApiService/handlerCreateFeedFollow"}
	for _, ctx := range []context.Context{
		context.Background(),
		metadata.NewIncomingContext(context.Background(), metadata.Pairs("api_key", "wrong")),
		metadata.NewIncomingContext(context.Background(), metadata.Pairs("api_key", "valid-test-key", "api_key", "second")),
	} {
		seen = false
		_, err := unaryAuth(ctx, nil, info, handler, testLookup)
		if status.Code(err) != codes.Unauthenticated || seen {
			t.Fatalf("unauthorized unary RPC reached handler: %v", err)
		}
		called := false
		err = streamAuth(nil, &testStream{ctx: ctx}, func(interface{}, grpc.ServerStream) error {
			called = true
			return nil
		}, testLookup)
		if status.Code(err) != codes.Unauthenticated || called {
			t.Fatalf("unauthorized stream reached handler: %v", err)
		}
	}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("api_key", "valid-test-key"))
	seen = false
	if _, err := unaryAuth(ctx, nil, info, handler, testLookup); err != nil || !seen {
		t.Fatalf("authorized unary RPC failed: %v", err)
	}
	if err := streamAuth(nil, &testStream{ctx: ctx}, func(_ interface{}, ss grpc.ServerStream) error {
		_, err := handler(ss.Context(), nil)
		return err
	}, testLookup); err != nil {
		t.Fatalf("authorized stream failed: %v", err)
	}
	health := &grpc.UnaryServerInfo{FullMethod: "/api.ApiService/handlerHealthz"}
	if _, err := unaryAuth(context.Background(), nil, health, func(context.Context, interface{}) (interface{}, error) {
		return nil, nil
	}, testLookup); err != nil {
		t.Fatalf("health probe required authentication: %v", err)
	}
}

func TestOwnershipAndMalformedIDsAreRejectedBeforeDatabaseAccess(t *testing.T) {
	ctx := auth.ContextWithUser(context.Background(), databaseUserToUser(database.User{
		ID: uuid.MustParse(ownerID), Name: "owner", ApiKey: "must-not-be-returned",
	}))
	server := &server{}
	for _, tc := range []struct {
		ctx   context.Context
		id    string
		limit string
		want  codes.Code
	}{
		{ctx, otherID, "12", codes.PermissionDenied},
		{ctx, "invalid", "12", codes.InvalidArgument},
		{ctx, ownerID, "101", codes.InvalidArgument},
		{context.Background(), ownerID, "12", codes.Unauthenticated},
	} {
		err := server.HandlerGetPostsForUser(&api.GetPostsForUserRequest{UserId: tc.id, Limit: tc.limit}, &testStream{ctx: tc.ctx})
		if status.Code(err) != tc.want {
			t.Fatalf("posts request: got %v, want %v", err, tc.want)
		}
	}
	if _, err := server.HandlerCreateFeedFollow(ctx, &api.CreateFeedFollowRequest{UserId: otherID, FeedId: otherID}); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("created follow for a different user: %v", err)
	}
	if _, err := server.HandlerCreateFeedFollow(ctx, &api.CreateFeedFollowRequest{UserId: ownerID, FeedId: "bad"}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("invalid feed ID was not rejected: %v", err)
	}
	if _, err := server.HandlerDeleteFeedFollows(ctx, &api.DeleteFeedFollowsRequest{UserId: ownerID, Id: "bad"}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("invalid follow ID was not rejected: %v", err)
	}
	user, err := server.HandlerGetUser(ctx, &api.GetUserRequest{})
	if err != nil || user.GetId() != ownerID || user.GetApiKey() != "" {
		t.Fatalf("user response contained the wrong identity or a key: %v", err)
	}
}
