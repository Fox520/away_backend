package main

import (
	"log"
	"net"
	"os"

	auth "github.com/Fox520/away_backend/auth"
	config "github.com/Fox520/away_backend/config"
	pb "github.com/Fox520/away_backend/user_service/pb"
	server "github.com/Fox520/away_backend/user_service/server"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc"
)

var logger = log.New(os.Stderr, "user_service_main: ", log.LstdFlags|log.Lshortfile)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Config load failed", err)
	}
	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		logger.Fatal("Failed to listen on port 9000", err)
	}

	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(auth.AuthInterceptor)),
		grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(auth.AuthInterceptor)))

	pb.RegisterUserServiceServer(grpcServer, server.NewUserServiceServer(cfg))

	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal("Failed to serve gRPC server", err)
	}

}
