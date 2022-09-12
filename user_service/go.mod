module github.com/Fox520/away_backend/user_service

go 1.16

replace github.com/Fox520/away_backend/testhelper => ../testhelper

require (
	cloud.google.com/go/iam v0.4.0 // indirect
	github.com/Fox520/away_backend/auth v0.0.0-20210820135334-717789c930c3
	github.com/Fox520/away_backend/testhelper v0.0.0-00010101000000-000000000000
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/google/uuid v1.3.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/jackc/pgx/v4 v4.17.2
	github.com/lib/pq v1.10.2
	github.com/olivere/elastic/v7 v7.0.29
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/onsi/gomega v1.20.2 // indirect
	github.com/spf13/viper v1.13.0
	golang.org/x/sys v0.0.0-20220909162455-aba9fc2a8ff2 // indirect
	google.golang.org/grpc v1.48.0
	google.golang.org/protobuf v1.28.1
)
