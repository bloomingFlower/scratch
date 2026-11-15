package auth

import (
	"context"
	"errors"
	api "github.com/bloomingFlower/rssagg/protos"
	"net/http"
	"strings"
)

// GetAPIKey extracts an API Key from
// the headers of an HTTP request
// Example:
// Authorization: APIKey {insert apikey here}
func GetAPIKey(headers http.Header) (string, error) {
	val := headers.Get("Authorization")
	if val == "" {
		return "", errors.New("no authentication info found")
	}

	vals := strings.Split(val, " ")
	if len(vals) != 2 {
		return "", errors.New("malformed auth header")

	}
	if vals[0] != "ApiKey" {
		return "", errors.New("malformed first part of auth header")
	}
	return vals[1], nil
}

// GetAPIKeyFromContext extracts an API Key from
// the context of a gRPC request
func GetAPIKeyFromContext(ctx context.Context) (string, error) {
	val := ctx.Value("apiKey")
	if val == nil {
		return "", errors.New("no API key found in context")
	}

	apiKey, ok := val.(string)
	if !ok {
		return "", errors.New("API key is not a string")
	}
	return apiKey, nil
}

type userContextKey struct{}

func ContextWithUser(ctx context.Context, user *api.User) context.Context {
	return context.WithValue(ctx, userContextKey{}, user)
}

func UserFromContext(ctx context.Context) (*api.User, bool) {
	user, ok := ctx.Value(userContextKey{}).(*api.User)
	return user, ok && user != nil
}
