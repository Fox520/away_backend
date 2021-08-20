package server

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	auth "github.com/Fox520/away_backend/auth"
	config "github.com/Fox520/away_backend/config"
	pb "github.com/Fox520/away_backend/user_service/pb"
	pq "github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var logger = log.New(os.Stderr, "user_service: ", log.LstdFlags|log.Lshortfile)

type UserServiceServer struct {
	pb.UnimplementedUserServiceServer
	DB *sql.DB
}

func NewUserServiceServer(cfg config.Config) *UserServiceServer {
	connectionString := fmt.Sprintf(`host=%s user=postgres password=%s dbname=%s port=%s sslmode=disable`,
		cfg.DBHost, cfg.DBPassword, cfg.DBName, cfg.DBPort)
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	logger.Print("Successfully connected to DB!")

	return &UserServiceServer{
		DB: db,
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

	sqlStatement := `
		INSERT INTO users (id, username, email, device_token, bio, verified, s_status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := server.DB.Exec(sqlStatement, userId, req.UserName, email, req.DeviceToken, req.Bio, false, "NONE")
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			// https://www.postgresql.org/docs/9.3/errcodes-appendix.html
			switch err.Code {
			case "23505": // unique_violation
				logger.Print("User create duplicate:", err.Code.Name())
				return nil, status.Error(codes.AlreadyExists, "User already exists")
			}
		}
		logger.Print("user insert error: ", err)
		return nil, status.Error(codes.Internal, "Could not create user")
	}

	return &pb.CreateUserResponse{Success: true}, nil
}

func (server *UserServiceServer) GetUser(ctx context.Context, ur *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	// If requested info belongs to that of requesting user, return full user object

	meta := ctx.Value(auth.ContextMetaDataKey).(map[string]string)
	userId := meta[auth.ContextUIDKey]
	var tempTime time.Time
	if userId == ur.Id {
		var user pb.AwayUser
		err := server.DB.QueryRow(`SELECT username, email, bio, device_token, verified, s_status, createdAt FROM users WHERE id = $1 LIMIT 1`, userId).Scan(
			&user.UserName,
			&user.Email,
			&user.Bio,
			&user.DeviceToken,
			&user.Verified,
			&user.SubscriptionStatus,
			&tempTime,
		)
		if err != nil {
			logger.Print("User not found: ", err)
			return nil, status.Error(codes.NotFound, "User not found")
		}
		user.CreatedAt = timestamppb.New(tempTime)
		user.Id = userId

		return &pb.GetUserResponse{
			UserOneof: &pb.GetUserResponse_User{User: &user},
		}, nil
	}
	var minimalUserInfo pb.MinimalUserInfo
	err := server.DB.QueryRow(`SELECT id, username, bio, verified, createdAt FROM users WHERE id = $1 LIMIT 1`, ur.Id).Scan(
		&minimalUserInfo.Id,
		&minimalUserInfo.UserName,
		&minimalUserInfo.Bio,
		&minimalUserInfo.Verified,
		&tempTime,
	)
	if err != nil {
		logger.Print("User not found [minimal]", err)
		return nil, status.Error(codes.NotFound, "User not found")
	}
	minimalUserInfo.CreatedAt = timestamppb.New(tempTime)
	return &pb.GetUserResponse{
		UserOneof: &pb.GetUserResponse_MinimalUser{MinimalUser: &minimalUserInfo},
	}, nil
}

func (server *UserServiceServer) UpdateUser(ctx context.Context, in *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	// Retrieve data from context
	meta := ctx.Value(auth.ContextMetaDataKey).(map[string]string)

	userId := meta[auth.ContextUIDKey]

	// Make sure certain fields are not empty
	if strings.ReplaceAll(in.UserName, " ", "") == "" {
		return nil, status.Error(codes.InvalidArgument, "Not all fields have data")
	}

	sqlStatement := `
		UPDATE users SET username = $1, bio = $2, device_token = $3 WHERE id = $4
	`
	_, err := server.DB.Exec(sqlStatement, in.UserName, in.Bio, in.DeviceToken, userId)
	if err != nil {
		logger.Print("update error: ", err)
		return nil, status.Error(codes.Internal, "Could not update user")
	}

	return &pb.UpdateUserResponse{Success: true}, nil
}

func (server *UserServiceServer) DeleteUser(ctx context.Context, dr *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {

	meta := ctx.Value(auth.ContextMetaDataKey).(map[string]string)

	userId := meta[auth.ContextUIDKey]

	deleteStmt := `DELETE FROM users WHERE id=$1`
	_, err := server.DB.Exec(deleteStmt, userId)

	if err != nil {
		return nil, err
	}
	return &pb.DeleteUserResponse{}, nil
}
