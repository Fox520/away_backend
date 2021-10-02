module github.com/Fox520/away_backend/property_service

go 1.17

require (
	github.com/Fox520/away_backend/auth v0.0.0-20211002101159-c54e7f2ed213
	github.com/Fox520/away_backend/config v0.0.0-20211002101159-c54e7f2ed213
	github.com/Fox520/away_backend/user_service v0.0.0-20211002083601-7bc74af9d4d6
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/lib/pq v1.10.3
	google.golang.org/grpc v1.41.0
	google.golang.org/protobuf v1.27.1
)

replace github.com/Fox520/away_backend/auth => ../auth

require (
	cloud.google.com/go v0.97.0 // indirect
	cloud.google.com/go/firestore v1.6.0 // indirect
	cloud.google.com/go/storage v1.17.0 // indirect
	firebase.google.com/go v3.13.0+incompatible // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/go-redis/redis/v8 v8.11.3 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/googleapis/gax-go/v2 v2.1.1 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mitchellh/mapstructure v1.4.2 // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.9.0 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	go.opencensus.io v0.23.0 // indirect
	golang.org/x/net v0.0.0-20210929193557-e81a3d93ecf6 // indirect
	golang.org/x/oauth2 v0.0.0-20210819190943-2bc19b11175f // indirect
	golang.org/x/sys v0.0.0-20211002104244-808efd93c36d // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/api v0.58.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20211001223012-bfb93cce50d9 // indirect
	gopkg.in/ini.v1 v1.63.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
