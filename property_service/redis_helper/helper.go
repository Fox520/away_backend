package redishelper

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	db "github.com/Fox520/away_backend/property_service/db"
	pb "github.com/Fox520/away_backend/property_service/pb"
)

const redisPropertyBase = "property:"
const redisMinimalPropertyBase = "minimal_property:"
const propertiesGeo string = "properties_geo"
const minimalPropertiesGeo string = "minimal_properties_geo"

func GetCachedSingleProperty(propertyId string) (*pb.Property, error) {
	if !shouldUseRedis() {
		return nil, errors.New("cannot connect to redis")
	}
	var prop pb.Property
	propertyCache, err := db.GetRedisClient().Get(redisPropertyBase + propertyId).Result()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(propertyCache), &prop)
	if err != nil {
		return nil, err
	}
	return &prop, nil
}

func CacheSingleProperty(prop *pb.Property) {
	if !shouldUseRedis() {
		return
	}
	propertyBytes, err := json.Marshal(prop)
	if err == nil {
		db.GetRedisClient().Set(redisPropertyBase+prop.Id, propertyBytes, 0).Err()
	}
}

func GetCachedMinimalProperty(propertyId string) (*pb.MinimalProperty, error) {
	if !shouldUseRedis() {
		return nil, errors.New("cannot connect to redis")
	}
	var prop pb.MinimalProperty
	propertyCache, err := db.GetRedisClient().Get(redisMinimalPropertyBase + propertyId).Result()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(propertyCache), &prop)
	if err != nil {
		return nil, err
	}
	return &prop, nil
}

func CacheMinimalProperty(prop *pb.MinimalProperty) {
	if !shouldUseRedis() {
		return
	}
	propertyBytes, err := json.Marshal(prop)
	if err == nil {
		db.GetRedisClient().Set(redisMinimalPropertyBase+prop.Id, propertyBytes, 0).Err()
	}
}

func FlushProperty(propertyId string) {
	if !shouldUseRedis() {
		return
	}
	db.GetRedisClient().ZRem(propertiesGeo, propertyId)
	db.GetRedisClient().ZRem(minimalPropertiesGeo, propertyId)
	db.GetRedisClient().Del(redisPropertyBase+propertyId, redisMinimalPropertyBase+propertyId)
}

var lastTimeRedisWasDown time.Time
var lastTimeRedisWasUp time.Time

// Checks the availability of redis every 3 minutes
// Should redis go down, requests will degrade for around 3 minutes
// Note: Adds TIME_OUTms to request if redis is unreachable
func shouldUseRedis() bool {
	if time.Since(lastTimeRedisWasDown).Seconds() < 180 {
		fmt.Println("redis is down* (3 min window)")
		return false
	}
	if time.Since(lastTimeRedisWasUp).Seconds() < 180 {
		return true
	}
	_, err := db.GetRedisClient().Ping().Result()
	if err != nil {
		lastTimeRedisWasDown = time.Now()
	} else {
		lastTimeRedisWasUp = time.Now()
	}
	return err == nil
}
