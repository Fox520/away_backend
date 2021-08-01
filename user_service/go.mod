module github.com/Fox520/away_backend/user_service

go 1.16

replace github.com/Fox520/away_backend/config => ../config

replace github.com/Fox520/away_backend/auth => ../auth

require (
	github.com/Fox520/away_backend/auth v0.0.0-00010101000000-000000000000
	github.com/Fox520/away_backend/config v0.0.0-00010101000000-000000000000
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/lib/pq v1.10.2
	github.com/spf13/cast v1.4.0 // indirect
	golang.org/x/net v0.0.0-20210726213435-c6fcb2dbf985 // indirect
	google.golang.org/genproto v0.0.0-20210729151513-df9385d47c1b // indirect
	google.golang.org/grpc v1.39.0
	google.golang.org/protobuf v1.27.1
)
