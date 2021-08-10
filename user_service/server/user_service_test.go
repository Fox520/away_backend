package server_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"testing"

	auth "github.com/Fox520/away_backend/auth"
	config "github.com/Fox520/away_backend/config"
	testhelper "github.com/Fox520/away_backend/testhelper"
	pb "github.com/Fox520/away_backend/user_service/pb"
	server "github.com/Fox520/away_backend/user_service/server"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

var lis *bufconn.Listener

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestMain(m *testing.M) {

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	// Overwrite since docker is on localhost
	cfg.DBHost = "localhost"

	// Setup database
	ctx := context.Background()
	container, port, _ := testhelper.CreateTestContainer(ctx, cfg)
	defer container.Terminate(ctx)

	// migration
	url := fmt.Sprintf("postgres://postgres:%s@localhost:%s/%s?sslmode=disable", cfg.DBPassword, port, cfg.DBName)
	db, _ := sql.Open("postgres", url)
	mig, _ := testhelper.NewPgMigrator(db)

	_ = mig.Up()

	lis = bufconn.Listen(1024 * 1024)

	// Overwrite the port with that of the container
	cfg.DBPort = port

	// Start server
	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(auth.AuthInterceptor)),
		grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(auth.AuthInterceptor)))

	pb.RegisterUserServiceServer(grpcServer, server.NewUserServiceServer(cfg))
	go func() {
		// In memory connections
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
	// Run the tests
	os.Exit(m.Run())
}

const mainUserEmail string = "testing@email.com"
const mainUserUID string = "Viism8HVdGfdhOcLUHEHqS7kA6m1"
const mainUserName string = "Alice"
const mainUserBio string = "main user bio"
const mainUserDeviceToken string = "12345"

const otherUserUID string = "zXiHzQfjmWf8Hte8L7RNEkNGf782"
const otherUserName string = "Jane"
const otherUserBio string = "other user bio"
const otherUserDeviceToken string = "abcde"

func TestCreateUser(t *testing.T) {
	conn, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := pb.NewUserServiceClient(conn)

	mainUserCtx := metadata.AppendToOutgoingContext(context.Background(), "token", testhelper.GetMainUserAuthToken())
	otherUserCtx := metadata.AppendToOutgoingContext(context.Background(), "token", testhelper.GetOtherUserAuthToken())

	t.Run("Create User Without All Fields", func(t *testing.T) {
		_, err = client.CreateUser(mainUserCtx, &pb.CreateUserRequest{})
		st, ok := status.FromError(err)
		if ok && st.Code() != codes.InvalidArgument {
			t.Fatal("User created without all fields")
		}
	})
	t.Run("Create User", func(t *testing.T) {
		_, err := client.CreateUser(mainUserCtx, &pb.CreateUserRequest{UserName: mainUserName, Bio: mainUserBio, DeviceToken: mainUserDeviceToken})
		if err != nil {
			t.Fatalf("CreateUser failed: %v", err)
		}
	})
	t.Run("Duplicate Create User", func(t *testing.T) {
		_, err = client.CreateUser(mainUserCtx, &pb.CreateUserRequest{UserName: mainUserName, Bio: mainUserBio, DeviceToken: mainUserDeviceToken})
		st, ok := status.FromError(err)
		if ok && st.Code() != codes.AlreadyExists {
			t.Fatal("Duplicate user created")
		}
	})

	t.Run("Create Secondary User ", func(t *testing.T) {
		_, err := client.CreateUser(otherUserCtx, &pb.CreateUserRequest{UserName: otherUserName, Bio: otherUserBio, DeviceToken: otherUserDeviceToken})
		if err != nil {
			t.Fatalf("CreateUser failed: %v", err)
		}
	})

}

func TestGetUser(t *testing.T) {
	conn, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())

	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := pb.NewUserServiceClient(conn)

	ctx := metadata.AppendToOutgoingContext(context.Background(), "token", testhelper.GetMainUserAuthToken())

	t.Run("Get non-existant user", func(t *testing.T) {
		_, err = client.GetUser(ctx, &pb.GetUserRequest{})
		st, ok := status.FromError(err)
		if ok && st.Code() != codes.NotFound {
			t.Fatal("Non-existant user found: " + fmt.Sprint(st.Code()))
		}
	})

	t.Run("Get own user", func(t *testing.T) {
		res, err := client.GetUser(ctx, &pb.GetUserRequest{Id: mainUserUID})
		if err != nil {
			t.Fatal(err)
		}

		switch u := res.UserOneof.(type) {
		case *pb.GetUserResponse_User:
			if u.User.Id != mainUserUID && u.User.Email != mainUserEmail && u.User.Bio != mainUserBio && u.User.UserName != mainUserName {
				t.Fatal("Fields do not match")
			}
		case *pb.GetUserResponse_MinimalUser:
			t.Fatal("Wrong message returned")
		}
	})

	t.Run("Get minimal user", func(t *testing.T) {
		res, err := client.GetUser(ctx, &pb.GetUserRequest{Id: otherUserUID})
		if err != nil {
			t.Fatal(err)
		}

		switch u := res.UserOneof.(type) {
		case *pb.GetUserResponse_User:
			t.Fatal("Wrong message returned")

		case *pb.GetUserResponse_MinimalUser:
			if u.MinimalUser.Id != otherUserUID && u.MinimalUser.UserName != otherUserName && u.MinimalUser.Bio != otherUserBio {
				t.Fatal("Fields do not match")
			}
		}
	})

}

func TestUpdateUser(t *testing.T) {
	conn, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())

	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := pb.NewUserServiceClient(conn)

	ctx := metadata.AppendToOutgoingContext(context.Background(), "token", testhelper.GetMainUserAuthToken())

	t.Run("Empty variables", func(t *testing.T) {
		_, err = client.UpdateUser(ctx, &pb.UpdateUserRequest{})
		st, ok := status.FromError(err)
		if ok && st.Code() != codes.InvalidArgument {
			t.Fatal("Empty variables went through: " + fmt.Sprint(st.Code()))
		}
	})

	t.Run("Actually update", func(t *testing.T) {
		_, err = client.UpdateUser(ctx, &pb.UpdateUserRequest{UserName: "sample", Bio: mainUserBio, DeviceToken: mainUserDeviceToken})
		if err != nil {
			t.Fatal("Could not update user")
		}
		// Check if new username reflects
		res, err := client.GetUser(ctx, &pb.GetUserRequest{Id: mainUserUID})
		if err != nil {
			t.Fatal(err)
		}

		switch u := res.UserOneof.(type) {
		case *pb.GetUserResponse_User:
			if u.User.UserName != "sample" {
				t.Fatal("Updated username does not reflect")
			}
		case *pb.GetUserResponse_MinimalUser:
			t.Fatal("Wrong message returned")
		}
	})

}

func TestDeleteUser(t *testing.T) {
	conn, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())

	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := pb.NewUserServiceClient(conn)

	ctx := metadata.AppendToOutgoingContext(context.Background(), "token", testhelper.GetMainUserAuthToken())

	t.Run("Delete user", func(t *testing.T) {
		_, err = client.DeleteUser(ctx, &pb.DeleteUserRequest{})
		if err != nil {
			t.Fatal("Could not delete user: " + err.Error())
		}
		_, err = client.GetUser(ctx, &pb.GetUserRequest{})
		st, ok := status.FromError(err)
		if ok && st.Code() != codes.NotFound {
			t.Fatal("Non-existant user found: " + fmt.Sprint(st.Code()))
		}
	})

}
