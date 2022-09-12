package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Fox520/away_backend/user_service/config"
	"github.com/go-redis/redis"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq"
	"github.com/olivere/elastic/v7"
)

var awayPoolOnce sync.Once
var awayPool *pgxpool.Pool
var awayDBOnce sync.Once
var awayDB *sql.DB
var redisClientOnce sync.Once
var redisClient *redis.Client

var elasticClient *elastic.Client
var elasticClientOnce sync.Once

func GetAwayPool() *pgxpool.Pool {
	awayPoolOnce.Do(func() {
		config := config.GetConfig()
		user := config.GetString("db.username")
		pwd := config.GetString("db.password")

		psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
			"password=%s dbname=away sslmode=disable pool_min_conns=1 pool_max_conns=%s",
			config.GetString("db.ip_address"),config.GetString("db.port"), user, pwd, config.GetString("db.pool_max_conns"))
		dbpool, err := pgxpool.Connect(context.Background(), psqlInfo)
		if err != nil {
			panic(err)
		}
		if err := dbpool.Ping(context.Background()); err != nil {
			panic(err)
		}
		awayPool = dbpool
	})
	return awayPool
}

func GetAwayDB() *sql.DB {
	awayDBOnce.Do(func() {
		config := config.GetConfig()
		user := config.GetString("db.username")
		pwd := config.GetString("db.password")

		psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
			"password=%s dbname=away sslmode=disable",
			config.GetString("db.ip_address"),config.GetString("db.port"), user, pwd)
		conn, err := sql.Open("postgres", psqlInfo)
		if err != nil {
			panic(err)
		}
		if err := conn.Ping(); err != nil {
			panic(err)
		}
		awayDB = conn
	})
	return awayDB
}

func GetRedisClient() *redis.Client {
	redisClientOnce.Do(func() {
		config := config.GetConfig()
		redisUrl := config.GetString("redis.url")

		redisClient = redis.NewClient(&redis.Options{
			Addr:        redisUrl,
			Password:    "",
			DB:          0,
			MaxRetries:  -1,
			DialTimeout: 400 * time.Millisecond,
		})
	})
	return redisClient
}

func GetElasticClient() *elastic.Client {
	elasticClientOnce.Do(func() {
		config := config.GetConfig()
		// client, err := elastic.NewClient(elastic.SetURL(cfg.ELASTICSEARCH_URL))
		client, err := elastic.NewClient(elastic.SetURL(config.GetString("elastic.url")))
		if err != nil {
			log.Println(err)
		}
		elasticClient = client

	})
	return elasticClient
}
