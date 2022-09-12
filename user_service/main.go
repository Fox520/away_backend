// $ go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
// $ protoc --go_out=./ --go-grpc_out=./ -I=../protos ../protos/user_service.proto
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	grpc_server "github.com/Fox520/away_backend/user_service/grpc_server"
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
