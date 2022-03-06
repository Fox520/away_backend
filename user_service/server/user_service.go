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
	pq "github.com/lib/pq"
	"github.com/olivere/elastic/v7"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var logger = log.New(os.Stderr, "user_service: ", log.LstdFlags|log.Lshortfile)

const ES_TIMEOUT time.Duration = 600 * time.Millisecond

// local tests show DB call is faster than hitting elastic
// Set to false when node is unreachable during startup
var USE_ELASTIC_SEARCH = true

type UserServiceServer struct {
	pb.UnimplementedUserServiceServer
	DB              *sql.DB
	Elastic         *elastic.Client
	UsersDbListener *pq.Listener
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
	// client, err := elastic.NewClient(elastic.SetURL(cfg.ELASTICSEARCH_URL))
	client, err := elastic.NewClient(elastic.SetURL("https://localhost:9200"))
	if err != nil {
		USE_ELASTIC_SEARCH = false
		logger.Println(err)
		logger.Println("Starting without Elasticsearch")
	}

	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	// Create listener for events which occur in the database
	usersListener := pq.NewListener(connectionString, 10*time.Second, time.Minute, reportProblem)
	err = usersListener.Listen("user_events")
	if err != nil {
		panic(err)
	}

	return &UserServiceServer{
		DB:              db,
		Elastic:         client,
		UsersDbListener: usersListener,
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

		err := server.DB.QueryRow(`
			SELECT users.username,
					users.email,
					users.bio,
					users.device_token,
					users.verified,
					users.s_status,
					users.createdat,
					users.profile_picture_url,
					subscription_status.s_description
			FROM   users, subscription_status
			WHERE  id = $1
					AND users.s_status = subscription_status.s_status
			LIMIT  1 
		`, userId).Scan(
			&user.UserName,
			&user.Email,
			&user.Bio,
			&user.DeviceToken,
			&user.Verified,
			&user.SubscriptionStatus,
			&tempTime,
			&user.ProfilePictureUrl,
			&user.SubscriptionStatusDescription,
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

func (server *UserServiceServer) StreamUserInfo(req *pb.StreamUserInfoRequest, stream pb.UserService_StreamUserInfoServer) error {
	meta := stream.Context().Value(auth.ContextMetaDataKey).(map[string]string)

	userId := meta[auth.ContextUIDKey]
	// Send initial data
	res, err := server.GetUser(stream.Context(), &pb.GetUserRequest{
		Id: userId,
	})
	if err != nil {
		return err
	}
	stream.Send(&pb.StreamUserInfoResponse{User: res.GetUser()})

	// We use this to avoid an extra db call
	// `UPDATE` event does not return previous value, thus store the original for later comparisons
	currentSubscriptionStatus := res.GetUser().SubscriptionStatus
	// NB: subscription status description is not in `users`, so events from the table will not include it
	// Store it in-memory for use in responses
	currentSubscriptionStatusDescription := res.GetUser().SubscriptionStatusDescription

	for {
		select {
		case n := <-server.UsersDbListener.Notify:

			output := streamResult{}
			if err := json.Unmarshal([]byte(n.Extra), &output); err != nil {
				logger.Println(err)
				return err
			}
			extractedUser := output.Data
			// Build the response user struct
			// We can't directly unmarshal due to timestamp conversion issues
			resultUser := pb.AwayUser{
				Id:                 extractedUser.Id,
				UserName:           extractedUser.UserName,
				Email:              extractedUser.Email,
				Bio:                extractedUser.Bio,
				DeviceToken:        extractedUser.DeviceToken,
				Verified:           extractedUser.Verified,
				SubscriptionStatus: extractedUser.SubscriptionStatus,
				// Retrieve status description from db if it changed
				SubscriptionStatusDescription: currentSubscriptionStatusDescription,
			}
			if resultUser.Id != userId {
				continue
			}
			// Now we can make that call and update in-memory version
			if resultUser.SubscriptionStatus != currentSubscriptionStatus {

				currentSubscriptionStatus = resultUser.SubscriptionStatus
				currentSubscriptionStatusDescription = resultUser.SubscriptionStatusDescription

				// todo: create getUser func with direct db access
				res, err := server.GetUser(stream.Context(), &pb.GetUserRequest{
					Id: userId,
				})
				if err != nil {
					return err
				}
				resultUser = *res.GetUser()
			}
			if err := stream.Send(&pb.StreamUserInfoResponse{
				User: &resultUser,
			}); err != nil {
				logger.Println(err)
				stream.Context().Done()
				return err
			}

		case <-time.After(90 * time.Second):
			// fmt.Println("Received no events for 90 seconds, checking connection")
			go func() {
				server.UsersDbListener.Ping()
			}()
		}
	}
}

type streamResult struct {
	Table  string        `json:"table"`
	Action string        `json:"action"`
	Data   streamUserObj `json:"data"`
}

// `createdAt` field is excluded for json decode to work; relevant error below
// parsing time "\"2022-03-06T16:32:03.258088\"" as "\"2006-01-02T15:04:05Z07:00\"": cannot parse "\"" as "Z07:00"
type streamUserObj struct {
	Id                 string `json:"id"`
	UserName           string `json:"username"`
	Email              string `json:"email"`
	Bio                string `json:"bio"`
	DeviceToken        string `json:"device_token"`
	Verified           bool   `json:"verified"`
	SubscriptionStatus string `json:"s_status"`
	ProfilePictureUrl  string `json:"profile_picture_url"`
}
