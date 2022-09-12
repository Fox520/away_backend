package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	auth "github.com/Fox520/away_backend/auth"
	"github.com/Fox520/away_backend/user_service/config"
	db "github.com/Fox520/away_backend/user_service/db/sqlc"
	pb "github.com/Fox520/away_backend/user_service/pb"
	userRepo "github.com/Fox520/away_backend/user_service/repository/user"
	pq "github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var logger = log.New(os.Stderr, "user_service: ", log.LstdFlags|log.Lshortfile)

const ES_TIMEOUT time.Duration = 600 * time.Millisecond

type UserServiceServer struct {
	pb.UnimplementedUserServiceServer
	usersDbListener *pq.Listener
	repo            userRepo.UserRepository
}

func NewUserServiceServer() *UserServiceServer {
	config := config.GetConfig()

	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			fmt.Println(err.Error())
		}
	}
	connectionString := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=away sslmode=disable ",
		config.GetString("db.ip_address"), config.GetString("db.port"), config.GetString("db.username"), config.GetString("db.password"))
	// Create listener for events which occur in the database
	usersListener := pq.NewListener(connectionString, 10*time.Second, time.Minute, reportProblem)
	// if err := usersListener.Ping(); err != nil {
	// 	panic(err)
	// }
	err := usersListener.Listen("user_events")
	if err != nil {
		panic(err)
	}

	return &UserServiceServer{
		repo:            *userRepo.NewUserRepository(),
		usersDbListener: usersListener,
	}
}

func (server *UserServiceServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {

	// Retrieve data from context
	meta := ctx.Value(auth.ContextMetaDataKey).(map[string]string)

	userId := meta[auth.ContextUIDKey]
	email := meta[auth.ContextEmailKey]

	// Make sure certain fields are not empty
	if strings.ReplaceAll(req.UserName, " ", "") == "" {
		return nil, status.Error(codes.InvalidArgument, "Not all fields have data")
	}
	pic := req.ProfilePictureUrl
	// Apple Sign In doesn't provide it again should the user re-register.
	if pic == "" {
		pic = "https://github.com/Fox520/Assets/blob/main/blank-profile-picture-g99784d5a8_640.png?raw=true"
	}
	_, err := server.repo.CreateUser(ctx, db.CreateUserParams{
		ID:                userId,
		Username:          req.UserName,
		Email:             email,
		DeviceToken:       req.DeviceToken,
		Bio:               req.Bio,
		Verified:          false,
		SStatus:           "NONE",
		ProfilePictureUrl: pic,
	})
	if err != nil {
		return nil, err
	}
	return &pb.CreateUserResponse{Success: true}, nil
}

func (server *UserServiceServer) GetUser(ctx context.Context, ur *pb.GetUserRequest) (*pb.GetUserResponse, error) {

	meta := ctx.Value(auth.ContextMetaDataKey).(map[string]string)
	userId := meta[auth.ContextUIDKey]

	// If requested info belongs to that of requesting user, return full user object
	if userId == ur.Id {
		result, err := server.repo.GetFullUser(ctx, ur.Id)
		if err != nil {
			return nil, err
		}
		user := pb.AwayUser{
			Id:                            result.ID,
			UserName:                      result.Username,
			Email:                         result.Email,
			Bio:                           result.Bio,
			DeviceToken:                   result.DeviceToken,
			Verified:                      result.Verified,
			SubscriptionStatus:            result.SStatus,
			SubscriptionStatusDescription: result.SDescription,
			ProfilePictureUrl:             result.ProfilePictureUrl,
			CreatedAt:                     timestamppb.New(result.Createdat),
		}

		return &pb.GetUserResponse{
			UserOneof: &pb.GetUserResponse_User{User: &user},
		}, nil
	}
	result, err := server.repo.GetMinimalUser(ctx, ur.Id)
	if err != nil {
		return nil, err
	}
	user := pb.MinimalUserInfo{
		Id:                result.ID,
		UserName:          result.Username,
		Bio:               result.Bio,
		Verified:          result.Verified,
		ProfilePictureUrl: result.ProfilePictureUrl,
		CreatedAt:         timestamppb.New(result.Createdat),
	}

	return &pb.GetUserResponse{
		UserOneof: &pb.GetUserResponse_MinimalUser{MinimalUser: &user},
	}, nil
}

func (server *UserServiceServer) UpdateUser(ctx context.Context, in *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	// Retrieve data from context
	meta := ctx.Value(auth.ContextMetaDataKey).(map[string]string)

	userId := meta[auth.ContextUIDKey]

	if in.UserName != "" && strings.ReplaceAll(in.UserName, " ", "") != "" {
		_, err := server.repo.SetUsername(ctx, db.SetUsernameParams{ID: userId, Username: in.UserName})
		if err != nil {
			return nil, err
		}
	}
	if in.Bio != "" && strings.ReplaceAll(in.Bio, " ", "") != "" {
		_, err := server.repo.SetBio(ctx, db.SetBioParams{ID: userId, Bio: in.Bio})
		if err != nil {
			return nil, err
		}
	}
	if in.DeviceToken != "" && strings.ReplaceAll(in.DeviceToken, " ", "") != "" {
		_, err := server.repo.SetDeviceToken(ctx, db.SetDeviceTokenParams{ID: userId, DeviceToken: in.DeviceToken})
		if err != nil {
			return nil, err
		}
	}
	if in.ProfilePictureUrl != "" && strings.ReplaceAll(in.ProfilePictureUrl, " ", "") != "" {
		_, err := server.repo.SetProfilePictureUrl(ctx, db.SetProfilePictureUrlParams{ID: userId, ProfilePictureUrl: in.ProfilePictureUrl})
		if err != nil {
			return nil, err
		}
	}

	return &pb.UpdateUserResponse{Success: true}, nil
}

func (server *UserServiceServer) DeleteUser(ctx context.Context, dr *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {

	meta := ctx.Value(auth.ContextMetaDataKey).(map[string]string)

	userId := meta[auth.ContextUIDKey]
	server.repo.DeleteUser(ctx, userId)
	return &pb.DeleteUserResponse{}, nil
}
