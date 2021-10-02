package auth

import (
	"context"

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

var FirebaseAuth = SetupFirebaseAuthClient()

// Retrieves and authenticates Firebase token from ctx
func AuthInterceptor(ctx context.Context) (context.Context, error) {
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
	_, err := FirebaseAuth.VerifyIDToken(context.Background(), token)
	if err != nil {
		logger.Print("Token verification error: ", err)
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	// Extract user id and email from the Firebase token then add to context

	uid, err := extractUserID(token)
	if err != nil {
		return nil, err
	}

	email, err := extractUserEmail(uid)
	if err != nil {
		return nil, err
	}
	payload := map[string]string{
		ContextEmailKey: email,
		ContextUIDKey:   uid,
		ContextTokenKey: token,
	}
	newCtx := context.WithValue(ctx, ContextMetaDataKey, payload)

	return newCtx, nil
}

func extractUserEmail(uid string) (string, error) {

	ur, err := FirebaseAuth.GetUser(context.Background(), uid)
	if err != nil {
		logger.Print("Cannot get user: ", err)
		return "", status.Error(codes.Unauthenticated, err.Error())
	}
	return ur.Email, nil
}

func extractUserID(token string) (string, error) {

	// https://firebase.google.com/docs/auth/admin/verify-id-tokens#go
	at, err := FirebaseAuth.VerifyIDToken(context.Background(), token)
	if err != nil {
		return "", status.Error(codes.Unauthenticated, err.Error())
	}
	return at.UID, nil
}