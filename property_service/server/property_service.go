package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
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
	"googlemaps.github.io/maps"
)

var logger = log.New(os.Stderr, "property_service: ", log.LstdFlags|log.Lshortfile)

const propertiesGeo string = "properties_geo"
const minimalPropertiesGeo string = "minimal_properties_geo"

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
	results, err := server.RedisClient.GeoRadius(stream.Context(), propertiesGeo, float64(req.Longitude), float64(req.Latitude), &redis.GeoRadiusQuery{
		Radius: float64(radius),
		Unit:   "m",
		Sort:   "ASC",
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
		server.RedisClient.GeoAdd(stream.Context(), propertiesGeo, &redis.GeoLocation{
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

	results, err := server.RedisClient.GeoRadius(stream.Context(), minimalPropertiesGeo, float64(req.Longitude), float64(req.Latitude), &redis.GeoRadiusQuery{
		Radius: float64(radius),
		Unit:   "m",
		Sort:   "ASC",
	}).Result()

	properties := make([]*pb.MinimalProperty, 0, 10)

	defer func() {
		properties = nil
	}()

	if err == nil {
		// Cache hit
		for _, geo := range results {
			prop, err := getCachedMinimalProperty(stream.Context(), server.RedisClient, geo.Name)
			if err != nil {
				// Cache miss, fetch & cache
				prop, err = fetchAndCacheMinimalProperty(stream.Context(), server.DB, server.RedisClient, geo.Name)
				if err != nil {
					continue
				}
			}
			properties = append(properties, prop)
		}
	}
	if len(properties) == 0 {
		properties, err = fetchAndCacheMinimalProperties(stream.Context(), server.DB, server.RedisClient, req.Latitude, req.Longitude, radius, false)
		if err != nil {
			return err
		}
	}
	var owners map[string]*user_pb.GetUserResponse
	for _, property := range properties {
		var singleMinimalProperty pb.SingleMinimalProperty
		// Avoid retrieving already fetched profiles
		if owners[property.UserID] == nil {
			// Attempt adding user profile
			token := meta[auth.ContextTokenKey]
			md := metadata.New(map[string]string{"token": token})
			requestContext := metadata.NewOutgoingContext(stream.Context(), md)
			res, err := server.UserClient.GetUser(requestContext, &user_pb.GetUserRequest{Id: property.UserID})
			if err == nil {
				singleMinimalProperty.Owner = res
			} else {
				logger.Println(err)
			}
		} else {
			singleMinimalProperty.Owner = owners[property.UserID]
		}
		singleMinimalProperty.Property = property
		// Cache property for this location
		server.RedisClient.GeoAdd(stream.Context(), minimalPropertiesGeo, &redis.GeoLocation{
			Longitude: float64(property.Longitude),
			Latitude:  float64(property.Latitude),
			Name:      property.Id,
		})

		stream.Send(&pb.GetMinimalPropertiesResponse{SingleMinimalProperty: &singleMinimalProperty})
	}

	return nil

}

/*
Just like GetMinimalInfoProperties but for promoted properties
*/
func (server *PropertyServiceServer) GetPromotedProperties(req *pb.PromotedRequest, stream pb.PropertyService_GetPromotedPropertiesServer) error {
	meta := stream.Context().Value(auth.ContextMetaDataKey).(map[string]string)

	radius := req.Radius
	if radius < 1 {
		radius = 5
	}
	// Convert radius (km) to meters
	radius = radius * 1000

	results, err := server.RedisClient.GeoRadius(stream.Context(), minimalPropertiesGeo, float64(req.Longitude), float64(req.Latitude), &redis.GeoRadiusQuery{
		Radius: float64(radius),
		Unit:   "m",
		Sort:   "ASC",
	}).Result()

	properties := make([]*pb.MinimalProperty, 0, 10)

	defer func() {
		properties = nil
	}()

	if err == nil {
		// Cache hit
		for _, geo := range results {
			prop, err := getCachedMinimalProperty(stream.Context(), server.RedisClient, geo.Name)
			if err != nil {
				// Cache miss, fetch & cache
				prop, err = fetchAndCacheMinimalProperty(stream.Context(), server.DB, server.RedisClient, geo.Name)
				if err != nil {
					continue
				}
			}
			// Only add those promoted
			if prop.Promoted {
				properties = append(properties, prop)
			}
		}
	}
	if len(properties) == 0 {
		properties, err = fetchAndCacheMinimalProperties(stream.Context(), server.DB, server.RedisClient, req.Latitude, req.Longitude, radius, true)
		if err != nil {
			return err
		}
	}
	var owners map[string]*user_pb.GetUserResponse
	for _, property := range properties {
		var singleMinimalProperty pb.SingleMinimalProperty
		// Avoid retrieving already fetched profiles
		if owners[property.UserID] == nil {
			// Attempt adding user profile
			token := meta[auth.ContextTokenKey]
			md := metadata.New(map[string]string{"token": token})
			requestContext := metadata.NewOutgoingContext(stream.Context(), md)
			res, err := server.UserClient.GetUser(requestContext, &user_pb.GetUserRequest{Id: property.UserID})
			if err == nil {
				singleMinimalProperty.Owner = res
			} else {
				logger.Println(err)
			}
		} else {
			singleMinimalProperty.Owner = owners[property.UserID]
		}
		singleMinimalProperty.Property = property
		// Cache property for this location
		server.RedisClient.GeoAdd(stream.Context(), minimalPropertiesGeo, &redis.GeoLocation{
			Longitude: float64(property.Longitude),
			Latitude:  float64(property.Latitude),
			Name:      property.Id,
		})

		stream.Send(&pb.PromotedResponse{Property: &singleMinimalProperty})
	}
	return nil
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
	// Clear from cache
	server.RedisClient.ZRem(ctx, propertiesGeo, dr.PropertyID)
	server.RedisClient.ZRem(ctx, minimalPropertiesGeo, dr.PropertyID)
	server.RedisClient.Del(ctx, redisPropertyBase+dr.PropertyID, redisMinimalPropertyBase+dr.PropertyID)
	return &pb.DeletePropertyResponse{
		Status: true,
	}, nil
}

func (server *PropertyServiceServer) GetFeaturedAreas(ctx context.Context, tr *pb.FeaturedAreasRequest) (*pb.FeaturedAreasResponse, error) {
	var response pb.FeaturedAreasResponse
	res, err := featuredAreasQuery(server.DB, tr.Country)
	if err != nil {
		return nil, err
	}
	response.FeaturedAreas = res
	return &response, nil
}

func (server *PropertyServiceServer) LocationSearch(stream pb.PropertyService_LocationSearchServer) error {
	sessionToken := maps.NewPlaceAutocompleteSessionToken()
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			logger.Println(err)
			return err
		}
		switch req := req.SearchOneof.(type) {
		case *pb.LocationSearchRequest_Query:
			predictions, err := performSearch(stream.Context(), sessionToken, &req.Query.Text, req.Query.CountryCode)
			if err != nil {
				return err
			}
			// Map each prediction to a LocationAutocomplete
			autocompletes := make([]*pb.LocationAutocomplete, len(predictions))
			for index, prediction := range predictions {
				autocompletes[index] = &pb.LocationAutocomplete{
					Title:         prediction.StructuredFormatting.MainText,
					SecondaryText: prediction.StructuredFormatting.SecondaryText,
					PlaceID:       prediction.PlaceID,
				}

			}
			// Send off the autocomplete slice
			response := pb.LocationSearchResponse{ResponseOneof: &pb.LocationSearchResponse_AutocompleteResponse{AutocompleteResponse: &pb.LocationAutocompleteResponse{
				Responses: autocompletes,
			}}}
			if err := stream.Send(&response); err != nil {
				return err
			}

		case *pb.LocationSearchRequest_Details:
			location, err := getLocation(stream.Context(), sessionToken, &req.Details.PlaceID)
			if err != nil {
				return err
			}
			stream.Send(&pb.LocationSearchResponse{
				ResponseOneof: &pb.LocationSearchResponse_Details{
					Details: &pb.LocationDetails{
						Latitude:  float32(location.Lat),
						Longitude: float32(location.Lng),
						PlaceID:   req.Details.PlaceID,
					},
				},
			})
			stream.Context().Done()
			return nil
		}

	}
}
