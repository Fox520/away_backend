package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	auth "github.com/Fox520/away_backend/auth"
	config "github.com/Fox520/away_backend/config"
	pb "github.com/Fox520/away_backend/user_service/github.com/Fox520/away_backend/user_service/pb"
	"github.com/go-redis/redis/v8"
	pq "github.com/lib/pq"
	"github.com/olivere/elastic/v7"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var logger = log.New(os.Stderr, "user_service: ", log.LstdFlags|log.Lshortfile)

const ES_TIMEOUT time.Duration = 600 * time.Millisecond

// local tests show DB call is faster than hitting elastic
const USE_ELASTIC_SEARCH = false

type UserServiceServer struct {
	pb.UnimplementedUserServiceServer
	DB          *sql.DB
	Elastic     *elastic.Client
	RedisClient *redis.Client
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
	rdb := redis.NewClient(&redis.Options{
		Addr:        "localhost:6379",
		Password:    "",
		DB:          0,
		MaxRetries:  -1,
		DialTimeout: 400 * time.Millisecond,
	})
	logger.Print("Successfully connected to DB!")
	client, err := elastic.NewSimpleClient(elastic.SetURL(cfg.ELASTICSEARCH_URL))
	if err != nil {
		logger.Println(err)
	}
	return &UserServiceServer{
		DB:          db,
		Elastic:     client,
		RedisClient: rdb,
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
	sqlStatement := `
		INSERT INTO users (id, username, email, device_token, bio, verified, s_status, profile_picture_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := server.DB.Exec(sqlStatement, userId, req.UserName, email, req.DeviceToken, req.Bio, false, "NONE", pic)
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

		if USE_ELASTIC_SEARCH {
			// Check if user is available in elastic search
			reqContext, cancel := context.WithTimeout(ctx, ES_TIMEOUT)
			defer cancel()
			get1, err := server.Elastic.Get().
				Index("users").
				Id(userId).
				Do(reqContext)

			if err == nil && get1.Found {
				if err = json.Unmarshal(get1.Source, &user); err == nil {
					return &pb.GetUserResponse{
						UserOneof: &pb.GetUserResponse_User{User: &user},
					}, nil
				}
				// Deserialization failed
			}
		}

		err := server.DB.QueryRow(`SELECT username, email, bio, device_token, verified, s_status, createdAt, profile_picture_url FROM users WHERE id = $1 LIMIT 1`, userId).Scan(
			&user.UserName,
			&user.Email,
			&user.Bio,
			&user.DeviceToken,
			&user.Verified,
			&user.SubscriptionStatus,
			&tempTime,
			&user.ProfilePictureUrl,
		)
		if err != nil {
			logger.Print("User not found: ", err)
			return nil, status.Error(codes.NotFound, "User not found")
		}
		user.CreatedAt = timestamppb.New(tempTime)
		user.Id = userId

		if USE_ELASTIC_SEARCH {
			// Add to ElasticSearch
			reqContext, cancel := context.WithTimeout(ctx, ES_TIMEOUT)
			defer cancel()
			_, err = server.Elastic.Index().
				Index("users").
				Id(user.Id).
				BodyJson(&user).
				Refresh("false").
				Do(reqContext)
			if err != nil {
				logger.Printf("Error: %s", err)
			}
		}

		return &pb.GetUserResponse{
			UserOneof: &pb.GetUserResponse_User{User: &user},
		}, nil
	}
	var minimalUserInfo pb.MinimalUserInfo

	if USE_ELASTIC_SEARCH {
		// Check if user is available in elastic search
		reqContext, cancel := context.WithTimeout(ctx, ES_TIMEOUT)
		defer cancel()
		get1, err := server.Elastic.Get().
			Index("minimal_users").
			Id(userId).
			Do(reqContext)
		if err == nil && get1.Found {
			if err = json.Unmarshal(get1.Source, &minimalUserInfo); err == nil {
				return &pb.GetUserResponse{
					UserOneof: &pb.GetUserResponse_MinimalUser{MinimalUser: &minimalUserInfo},
				}, nil
			}
			// Deserialization failed
		}
	}

	err := server.DB.QueryRow(`SELECT id, username, bio, verified, createdAt, profile_picture_url FROM users WHERE id = $1 LIMIT 1`, ur.Id).Scan(
		&minimalUserInfo.Id,
		&minimalUserInfo.UserName,
		&minimalUserInfo.Bio,
		&minimalUserInfo.Verified,
		&tempTime,
		&minimalUserInfo.ProfilePictureUrl,
	)
	if err != nil {
		logger.Print("User not found [minimal]", err)
		return nil, status.Error(codes.NotFound, "User not found")
	}
	minimalUserInfo.CreatedAt = timestamppb.New(tempTime)

	if USE_ELASTIC_SEARCH {
		// Add to ElasticSearch
		reqContext, cancel := context.WithTimeout(ctx, ES_TIMEOUT)
		defer cancel()
		_, err = server.Elastic.Index().
			Index("minimal_users").
			Id(minimalUserInfo.Id).
			BodyJson(&minimalUserInfo).
			Refresh("false").
			Do(reqContext)
		if err != nil {
			logger.Printf("Error: %s", err)
		}

	}

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
		UPDATE users SET username = $1, bio = $2, device_token = $3, profile_picture_url = $4 WHERE id = $5
	`
	_, err := server.DB.Exec(sqlStatement, in.UserName, in.Bio, in.DeviceToken, in.ProfilePictureUrl, userId)
	if err != nil {
		logger.Print("update error: ", err)
		return nil, status.Error(codes.Internal, "Could not update user")
	}
	if USE_ELASTIC_SEARCH {
		reqContext, cancel := context.WithTimeout(ctx, ES_TIMEOUT)
		defer cancel()
		server.Elastic.Delete().Index("users").Id(userId).Do(reqContext)
		server.Elastic.Delete().Index("minimal_users").Id(userId).Do(reqContext)
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
	if err = auth.FirebaseAuth.DeleteUser(context.Background(), userId); err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	if USE_ELASTIC_SEARCH {
		reqContext, cancel := context.WithTimeout(ctx, ES_TIMEOUT)
		defer cancel()
		server.Elastic.Delete().Index("users").Id(userId).Do(reqContext)
		server.Elastic.Delete().Index("minimal_users").Id(userId).Do(reqContext)
	}

	return &pb.DeleteUserResponse{}, nil
}
