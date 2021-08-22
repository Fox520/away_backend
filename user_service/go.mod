module github.com/Fox520/away_backend/user_service

go 1.16

replace github.com/Fox520/away_backend/config => ../config

replace github.com/Fox520/away_backend/auth => ../auth

replace github.com/Fox520/away_backend/testhelper => ../testhelper

require (
	github.com/Fox520/away_backend/auth v0.0.0-20210820135334-717789c930c3
	github.com/Fox520/away_backend/config v0.0.0-20210807233659-2b4e69eac5fd
	github.com/Fox520/away_backend/testhelper v0.0.0-20210820135334-717789c930c3
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/lib/pq v1.10.2
	github.com/spf13/cast v1.4.0 // indirect
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.27.1
)
