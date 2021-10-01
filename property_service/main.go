// protoc --go_out=pb --go-grpc_out=pb protos/user_service.proto
package main

import (
	"log"
	"net"
	"os"

	auth "github.com/Fox520/away_backend/auth"
	config "github.com/Fox520/away_backend/config"
	pb "github.com/Fox520/away_backend/property_service/pb"
	server "github.com/Fox520/away_backend/property_service/server"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc"
)

var logger = log.New(os.Stderr, "property_service_main: ", log.LstdFlags|log.Lshortfile)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Config load failed", err)
	}

	lis, err := net.Listen("tcp", ":9001")
	if err != nil {
		log.Fatalf("Failed to listen on port 9001: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(auth.AuthInterceptor)),
		grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(auth.AuthInterceptor)))

	pb.RegisterPropertyServiceServer(grpcServer, server.NewPropertyServiceServer(cfg))

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}

}
