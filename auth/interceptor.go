package auth

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Required to pass data through context
type key string

const ContextMetaDataKey key = "metadata"
const ContextEmailKey string = "auth.email"
const ContextUIDKey string = "auth.uid"
const ContextTokenKey string = "token"

// Retrieves and authenticates Firebase token from ctx
func EnsureFirebaseToken(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		logger.Print("No metadata found")
		return nil, status.New(codes.NotFound, "payload missing").Err()
	}
	tokenSlice, ok := md["token"]

	if !ok {
		logger.Print("Token not found")
		return nil, status.New(codes.NotFound, "token not found").Err()
	}
	token := tokenSlice[0]
	// Uncomment for token verification to take place
	// authToken, err := GetFirebaseAuthClient().VerifyIDToken(context.Background(), token)
	authToken, err := GetFirebaseAuthClient().VerifyIDToken(context.Background(), token)
	if err != nil {
		logger.Print("Token verification error: ", err)
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	// identities => {"email":[name@example.com]}
	emails := authToken.Firebase.Identities["email"]
	email := ""
	switch t := emails.(type) {
	case []interface{}:
		email = fmt.Sprint(t[0])
	default:
		return nil, status.Error(codes.Internal, "Email not found")
	}

	payload := map[string]string{
		ContextEmailKey: email,
		ContextUIDKey:   authToken.UID,
		ContextTokenKey: token,
	}
	newCtx := context.WithValue(ctx, ContextMetaDataKey, payload)

	return newCtx, nil
}
