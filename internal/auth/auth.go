package auth

import (
	"context"
	"errors"
	api "github.com/bloomingFlower/rssagg/protos"
	"log"
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
	log.Printf("Get API key: %v", ctx)

	if val == nil {
		return "", errors.New("no API key found in context")
	}

	apiKey, ok := val.(string)
	if !ok {
		return "", errors.New("API key is not a string")
	}
	return apiKey, nil
}

func ContextWithUser(ctx context.Context, user *api.User) context.Context {
	return context.WithValue(ctx, "user", user)
}
