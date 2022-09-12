// protoc --go_out=.\\ --go-grpc_out=.\\ -I=protos -I=..\\user_service\\protos  protos/property_service.proto
package main

// import (
// 	"log"
// 	"net"
// 	"os"

// 	auth "github.com/Fox520/away_backend/auth"
// 	config "github.com/Fox520/away_backend/config"
// 	pb "github.com/Fox520/away_backend/property_service/github.com/Fox520/away_backend/property_service/pb"
// 	server "github.com/Fox520/away_backend/property_service/server"
// 	user_pb "github.com/Fox520/away_backend/user_service/pb"
// 	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
// 	"github.com/oxequa/grace"
// 	"google.golang.org/grpc"
// )

// var logger = log.New(os.Stderr, "property_service_main: ", log.LstdFlags|log.Lshortfile)

// func main() {
// 	cfg, err := config.LoadConfig()
// 	if err != nil {
// 		logger.Fatal("Config load failed", err)
// 	}

// 	lis, err := net.Listen("tcp", ":9001")
// 	if err != nil {
// 		log.Fatalf("Failed to listen on port 9001: %v", err)
// 	}

// 	grpcServer := grpc.NewServer(
// 		grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(auth.EnsureFirebaseToken)),
// 		grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(auth.EnsureFirebaseToken)))

// 	// User service connection
// 	conn, err := grpc.Dial("localhost:9000", grpc.WithInsecure())

// 	var userClient user_pb.UserServiceClient
// 	if err != nil {
// 		// Service should still operate even if user service is unreachable
// 		logger.Printf("Failed to connect to user service: %s", err)
// 	} else {
// 		defer conn.Close()
// 		userClient = user_pb.NewUserServiceClient(conn)
// 	}
// 	pb.RegisterPropertyServiceServer(grpcServer, server.NewPropertyServiceServer(cfg, userClient))
// 	// Upon panic, try restarting. Max 3 times
// 	var restartAttempts int = 0
// 	for {
// 		err = start(grpcServer, lis)
// 		if err == nil {
// 			break
// 		}
// 		if restartAttempts < 3 {
// 			restartAttempts += 1
// 			logger.Println("Restarting...")
// 		} else {
// 			logger.Println("Shutting down")
// 			break
// 		}
// 	}

// }

// func start(grpcServer *grpc.Server, lis net.Listener) (e error) {
// 	defer grace.Recover(&e)
// 	if err := grpcServer.Serve(lis); err != nil {
// 		log.Fatalf("Failed to serve gRPC server: %v", err)
// 	}
// 	return
// }

import (
	"flag"
	"fmt"
	"log"
	"os"

	grpc_server "github.com/Fox520/away_backend/property_service/grpc_server"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	environment := os.Getenv("DEPLOYMENT_ENV")
	if environment == "" {
		environment = "development"
	}
	log.Println("Running in: ", environment)
	flag.Usage = func() {
		fmt.Println("Usage: server -e {mode}")
		os.Exit(1)
	}
	flag.Parse()
	grpc_server.Init()
}
