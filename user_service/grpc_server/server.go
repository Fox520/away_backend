package grpc_server

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/Fox520/away_backend/auth"
	"github.com/Fox520/away_backend/user_service/config"
	"github.com/Fox520/away_backend/user_service/pb"
	"github.com/Fox520/away_backend/user_service/server"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc"
)

// protoc --go_out=./ --go-grpc_out=./ -I=../protos ../protos/user_service.proto
func Init() {
	config := config.GetConfig()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", config.GetString("grpc_server.port")))
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to listen on port %s", config.GetString("grpc_server.port")), err)
	}
	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(auth.EnsureFirebaseToken)),
		grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(auth.EnsureFirebaseToken)))

	pb.RegisterUserServiceServer(grpcServer, server.NewUserServiceServer())

	go func() {
		// service connections
		log.Println("listening")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal("Failed to serve gRPC server", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Graceful shutdown...")

	grpcServer.GracefulStop()
	log.Println("Server stopped")
}
