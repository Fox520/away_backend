package grpc_server

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/Fox520/away_backend/auth"
	config "github.com/Fox520/away_backend/property_service/config"

	"github.com/Fox520/away_backend/property_service/pb"
	user_pb "github.com/Fox520/away_backend/user_service/pb"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc"
)

func Init() {
	conf := config.GetConfig()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", conf.GetString("grpc_server.port")))
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to listen on port %s", conf.GetString("grpc_server.port")), err)
	}
	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(auth.EnsureFirebaseToken)),
		grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(auth.EnsureFirebaseToken)))

	var userClient user_pb.UserServiceClient
	conn, err := grpc.Dial(conf.GetString("user_service.address"), grpc.WithInsecure())

	if err != nil {
		// Service should still operate even if user service is unreachable
		log.Printf("Failed to connect to user service: %s", err)
	} else {
		defer conn.Close()
		userClient = user_pb.NewUserServiceClient(conn)
	}
	pb.RegisterPropertyServiceServer(grpcServer, NewPropertyServiceServer(userClient))

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
