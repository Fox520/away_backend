package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	auth "github.com/Fox520/away_backend/auth"
	config "github.com/Fox520/away_backend/config"
	pb "github.com/Fox520/away_backend/property_service/github.com/Fox520/away_backend/property_service/pb"
	user_pb "github.com/Fox520/away_backend/user_service/pb"
	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var logger = log.New(os.Stderr, "property_service: ", log.LstdFlags|log.Lshortfile)

type PropertyServiceServer struct {
	pb.UnimplementedPropertyServiceServer
	DB          *sql.DB
	UserClient  user_pb.UserServiceClient
	RedisClient *redis.Client
}

func NewPropertyServiceServer(cfg config.Config, userClient user_pb.UserServiceClient) *PropertyServiceServer {
	connectionString := fmt.Sprintf(`host=%s user=postgres password=%s dbname=%s port=%s sslmode=disable`,
		cfg.DBHost, cfg.DBPassword, cfg.DBName, cfg.DBPort)
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:        "localhost:6379",
		Password:    "",
		DB:          0,
		MaxRetries:  -1,
		DialTimeout: 400 * time.Millisecond,
	})
	logger.Println("Successfully connected to DB!")

	return &PropertyServiceServer{
		DB:          db,
		UserClient:  userClient,
		RedisClient: rdb,
	}

}

// Fetch property from db
func (server *PropertyServiceServer) GetSingleProperty(ctx context.Context, pr *pb.GetSinglePropertyRequest) (*pb.SinglePropertyResponse, error) {

	var propertyResponse pb.SinglePropertyResponse
	// Fixes: invalid memory address or nil pointer dereference
	propertyResponse.Property = &pb.Property{}

	// Check for cache hit
	prop, err := getCachedSingleProperty(ctx, server.RedisClient, pr.Id)

	if err != nil {
		pro, err := fetchAndCacheSingleProperty(ctx, server.DB, server.RedisClient, pr.Id)
		if err != nil {
			return nil, err
		}
		propertyResponse.Property = pro
	} else {
		propertyResponse.Property = prop
	}

	// Attempt adding user profile
	meta := ctx.Value(auth.ContextMetaDataKey).(map[string]string)

	token := meta[auth.ContextTokenKey]
	md := metadata.New(map[string]string{"token": token})
	requestContext := metadata.NewOutgoingContext(ctx, md)
	res, err := server.UserClient.GetUser(requestContext, &user_pb.GetUserRequest{Id: propertyResponse.Property.UserID})
	if err == nil {
		propertyResponse.Owner = res
	} else {
		logger.Println(err)
	}
	return &propertyResponse, nil
}

func (server *PropertyServiceServer) GetMultipleProperties(req *pb.GetMultiplePropertyRequest, stream pb.PropertyService_GetMultiplePropertiesServer) error {
	meta := stream.Context().Value(auth.ContextMetaDataKey).(map[string]string)

	radius := req.Radius
	if radius < 1 {
		radius = 5
	}
	// Convert radius (km) to meters
	radius = radius * 1000
	results, err := server.RedisClient.GeoRadius(stream.Context(), "properties_geo", float64(req.Longitude), float64(req.Latitude), &redis.GeoRadiusQuery{
		Radius: float64(radius),
		Unit:   "m",
	}).Result()

	properties := make([]*pb.Property, 0, 10)

	defer func() {
		properties = nil
	}()

	if err == nil {
		// Cache hit
		for _, geo := range results {
			prop, err := getCachedSingleProperty(stream.Context(), server.RedisClient, geo.Name)
			if err != nil {
				// Cache miss, fetch & cache
				prop, err = fetchAndCacheSingleProperty(stream.Context(), server.DB, server.RedisClient, geo.Name)
				if err != nil {
					continue
				}
			}
			properties = append(properties, prop)
		}
	}
	if len(properties) == 0 {
		// src https://dba.stackexchange.com/a/158422
		propertyRows, err := server.DB.Query(`
			SELECT
				p.id,
				p.user_id,
				ptype.p_type,
				ptype.id,
				pcat.p_category,
				pcat.id,
				pusage.p_usage,
				pusage.id,
				p.bedrooms,
				p.bathrooms,
				p.surburb,
				p.town,
				p.title,
				p.p_description,
				p.currency,
				p.available,
				p.price,
				p.deposit,
				p.sharing_price,
				p.pets_allowed,
				p.free_wifi,
				p.water_included,
				p.electricity_included,
				p.latitude,
				p.longitude,
				p.posted_date
			FROM
				properties p,
				lateral(SELECT id, p_type FROM property_type WHERE id = p.property_type_id) as ptype,
				lateral(SELECT id, p_category FROM property_category WHERE id = p.property_category_id) as pcat,
				lateral(SELECT id, p_usage FROM property_usage WHERE id = p.property_usage_id) as pusage
			WHERE
				/* First condition allows to search for points at an approximate distance:
				a distance computed using a 'box', instead of a 'circumference'.
				This first condition will use the index.
				(45.1013021, 46.3021011) = (lat, lng) of search center. 
				25000 = search radius (in m)
				*/
				earth_box(ll_to_earth($1, $2), $3) @> ll_to_earth(latitude, longitude) 

				/* This second condition (which is slower) will "refine" 
				the previous search, to include only the points within the
				circumference.
				*/
				AND earth_distance(ll_to_earth($1, $2), 
						ll_to_earth(latitude, longitude)) < $3 
		`,
			req.Latitude,
			req.Longitude,
			radius,
		)

		if err != nil {
			return err
		}
		for propertyRows.Next() {
			var property pb.Property
			var tempTime time.Time

			err = propertyRows.Scan(&property.Id,
				&property.UserID,
				&property.PropertyType,
				&property.PropertyTypeID,
				&property.PropertyCategory,
				&property.PropertyCategoryID,
				&property.PropertyUsage,
				&property.PropertyUsageID,
				&property.Bedrooms,
				&property.Bathrooms,
				&property.Surburb,
				&property.Town,
				&property.Title,
				&property.Description,
				&property.Currency,
				&property.Available,
				&property.Price,
				&property.Deposit,
				&property.SharingPrice,
				&property.PetsAllowed,
				&property.FreeWifi,
				&property.WaterIncluded,
				&property.ElectricityIncluded,
				&property.Latitude,
				&property.Longitude,
				&tempTime)
			if err != nil {
				logger.Println(err)
				continue
			}
			property.PostedDate = timestamppb.New(tempTime)

			// Add photos to response
			rows, err := server.DB.Query(`SELECT id, p_url, property_id FROM property_photos WHERE property_id = $1`, property.Id)
			if err != nil {
				continue
			}

			for rows.Next() {
				id := ""
				url := ""
				propId := ""
				err = rows.Scan(&id, &url, &propId)
				if err != nil {
					continue
				}
				property.Photos = append(property.Photos, &pb.Photo{Id: id, Url: url, PropertyID: propId})
			}
			properties = append(properties, &property)
			cacheSingleProperty(stream.Context(), server.RedisClient, &property)
		}
	}
	var owners map[string]*user_pb.GetUserResponse
	for _, property := range properties {
		var response pb.SinglePropertyResponse
		// Avoid retrieving already fetched profiles
		if owners[property.UserID] == nil {
			// Attempt adding user profile
			token := meta[auth.ContextTokenKey]
			md := metadata.New(map[string]string{"token": token})
			requestContext := metadata.NewOutgoingContext(stream.Context(), md)
			res, err := server.UserClient.GetUser(requestContext, &user_pb.GetUserRequest{Id: property.UserID})
			if err == nil {
				response.Owner = res
			} else {
				logger.Println(err)
			}
		} else {
			response.Owner = owners[property.UserID]
		}
		response.Property = property
		// Cache property
		server.RedisClient.GeoAdd(stream.Context(), "properties_geo", &redis.GeoLocation{
			Longitude: float64(property.Longitude),
			Latitude:  float64(property.Latitude),
			Name:      property.Id,
		})

		stream.Send(&pb.GetMultiplePropertyResponse{Response: &response})
	}

	return nil
}

// req *pb.GetMultiplePropertyRequest, stream pb.PropertyService_GetMultiplePropertiesServer
func (server *PropertyServiceServer) GetMinimalInfoProperties(req *pb.GetMinimalPropertiesRequest, stream pb.PropertyService_GetMinimalInfoPropertiesServer) error {
	meta := stream.Context().Value(auth.ContextMetaDataKey).(map[string]string)

	radius := req.Radius
	if radius < 1 {
		radius = 5
	}
	// Convert radius (km) to meters
	radius = radius * 1000

	properties, err := minimalQuery(server.DB, req.Latitude, req.Longitude, radius, false)
	if err != nil {
		return err
	}

	token := meta[auth.ContextTokenKey]
	md := metadata.New(map[string]string{"token": token})
	requestContext := metadata.NewOutgoingContext(stream.Context(), md)
	for _, prop := range properties {
		// Attempt adding user profile
		userResponse, err := server.UserClient.GetUser(requestContext, &user_pb.GetUserRequest{Id: prop.UserID})
		if err != nil {
			logger.Println(err)
		}
		stream.Send(&pb.GetMinimalPropertiesResponse{SingleMinimalProperty: &pb.SingleMinimalProperty{
			Property: prop,
			Owner:    userResponse,
		}})
	}

	return nil

}

func (server *PropertyServiceServer) GetPromotedProperties(ctx context.Context, pr *pb.PromotedRequest) (*pb.PromotedResponse, error) {
	radius := pr.Radius
	if radius < 1 {
		radius = 200
	}
	// Convert radius (km) to meters
	radius = radius * 1000

	var promotedResponse pb.PromotedResponse
	properties, err := minimalQuery(server.DB, pr.Latitude, pr.Longitude, radius, true)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	meta := ctx.Value(auth.ContextMetaDataKey).(map[string]string)

	token := meta[auth.ContextTokenKey]
	md := metadata.New(map[string]string{"token": token})
	requestContext := metadata.NewOutgoingContext(ctx, md)
	for _, prop := range properties {
		// Attempt adding user profile
		userResponse, err := server.UserClient.GetUser(requestContext, &user_pb.GetUserRequest{Id: prop.UserID})
		if err != nil {
			logger.Println(err)
		}
		promotedResponse.Properties = append(promotedResponse.Properties, &pb.SingleMinimalProperty{
			Property: prop,
			Owner:    userResponse,
		})
	}

	return &promotedResponse, nil
}

func (server *PropertyServiceServer) CreateProperty(ctx context.Context, request *pb.CreatePropertyRequest) (*pb.Property, error) {
	property := request.Property
	if property.Currency == "" || property.Title == "" || property.Photos == nil {
		return nil, errors.New("field requirements not met")
	}
	meta := ctx.Value(auth.ContextMetaDataKey).(map[string]string)
	userId := meta[auth.ContextUIDKey]

	property.UserID = userId

	sqlStatement := `
		INSERT INTO properties
		(user_id,
		property_type_id,
		property_category_id,
		property_usage_id,
		bedrooms,
		bathrooms,
		surburb,
		town,
		title,
		p_description,
		currency,
		available,
		price,
		deposit,
		sharing_price,
		pets_allowed,
		free_wifi,
		water_included,
		electricity_included,
		latitude,
		longitude)
	VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14,
		$15, $16, $17, $18, $19, $20, $21)
	RETURNING id
	`
	propertyId := ""
	err := server.DB.QueryRow(sqlStatement,
		property.UserID,
		property.PropertyTypeID,
		property.PropertyCategoryID,
		property.PropertyUsageID,
		property.Bedrooms,
		property.Bathrooms,
		property.Surburb,
		property.Town,
		property.Title,
		property.Description,
		property.Currency,
		property.Available,
		property.Price,
		property.Deposit,
		property.SharingPrice,
		property.PetsAllowed,
		property.FreeWifi,
		property.WaterIncluded,
		property.ElectricityIncluded,
		property.Latitude,
		property.Longitude,
	).Scan(&propertyId)
	if err != nil {
		return nil, err
	}

	// Now add the photos
	for _, photo := range property.Photos {
		sqlStatement := `
				INSERT INTO property_photos (property_id, p_url)
				VALUES ($1, $2)
			`
		_, err = server.DB.Exec(sqlStatement, propertyId, photo.Url)
		if err != nil {
			// TODO: Delete property as photos failed to add
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	property.Id = propertyId
	// Retrieve property
	resp, err := server.GetSingleProperty(ctx, &pb.GetSinglePropertyRequest{Id: propertyId})
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	return resp.Property, nil
}

func (server *PropertyServiceServer) UpdateProperty(context.Context, *pb.Property) (*pb.Property, error) {
	return nil, nil
}

func (server *PropertyServiceServer) DeleteProperty(ctx context.Context, dr *pb.DeletePropertyRequest) (*pb.DeletePropertyResponse, error) {
	meta := ctx.Value(auth.ContextMetaDataKey).(map[string]string)

	userId := meta[auth.ContextUIDKey]
	propOwnerID := ""
	// Make sure it exists and belongs to requesting owner
	err := server.DB.QueryRow(`
		SELECT user_id FROM properties WHERE id = $1`, dr.PropertyID).Scan(&propOwnerID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "Property not found")
	}
	if userId != propOwnerID {
		return nil, status.New(codes.PermissionDenied, "unauthorised request").Err()
	}
	deleteStmt := `DELETE FROM properties WHERE id=$1`
	_, err = server.DB.Exec(deleteStmt, userId)

	if err != nil {
		return nil, err
	}
	return &pb.DeletePropertyResponse{
		Status: true,
	}, nil
}

func (server *PropertyServiceServer) GetFeaturedAreas(ctx context.Context, tr *pb.FeaturedAreasRequest) (*pb.FeaturedAreasResponse, error) {
	var response pb.FeaturedAreasResponse
	res, err := topAreasQuery(server.DB, tr.Country)
	if err != nil {
		return nil, err
	}
	response.FeaturedAreas = res
	return &response, nil
}
