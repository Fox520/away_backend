package property

import (
	"context"
	"log"

	"github.com/go-redis/redis"

	"github.com/Fox520/away_backend/property_service/pb"
	redis_helper "github.com/Fox520/away_backend/property_service/redis_helper"

	db_collections "github.com/Fox520/away_backend/property_service/db"
	db "github.com/Fox520/away_backend/property_service/db/sqlc"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// these should only be in redis helper; remove here
const propertiesGeo string = "properties_geo"
const minimalPropertiesGeo string = "minimal_properties_geo"

type PropertyRepository struct {
	query *db.Queries
}

func NewPropertyRepository() *PropertyRepository {
	return &PropertyRepository{
		query: db.New(db_collections.GetAwayDB()),
	}
}

func (repo *PropertyRepository) CreateProperty(ctx context.Context, arg db.CreatePropertyParams) (*db.Property, error) {
	property, err := repo.query.CreateProperty(ctx, arg)

	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			// https://www.postgresql.org/docs/9.3/errcodes-appendix.html
			switch err.Code {
			case "23505": // unique_violation
				log.Println("User create duplicate:", err.Code.Name())
				return nil, status.Error(codes.AlreadyExists, "Property already exists")
			}
		}
		log.Println("property insert error: ", err)
		return nil, status.Error(codes.InvalidArgument, "Could not create property")
	}
	return &property, nil
}

func (repo *PropertyRepository) CreatePropertyPhoto(ctx context.Context, arg db.CreatePropertyPhotoParams) (*db.PropertyPhoto, error) {
	result, err := repo.query.CreatePropertyPhoto(ctx, arg)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &result, nil
}

func (repo *PropertyRepository) GetProperty(ctx context.Context, propertyId uuid.UUID) (*pb.Property, error) {
	// Check for cache hit
	cachedProp, err := redis_helper.GetCachedSingleProperty(propertyId.String())
	if err == nil {
		return cachedProp, nil
	}
	result, err := repo.query.GetProperty(ctx, propertyId)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			// https://www.postgresql.org/docs/9.3/errcodes-appendix.html
			switch err.Code {
			case "02000": // no_data
				return nil, status.Error(codes.NotFound, "Property not found")
			}
		}
		log.Println(err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	photos, err := repo.GetPropertyPhotos(ctx, propertyId) // fetchAndCacheSingleProperty(ctx, pr.Id)
	if err != nil {
		return nil, err
	}
	property := pb.Property{
		Id:                  result.ID.String(),
		UserID:              result.UserID,
		PropertyTypeID:      int32(result.PropertyTypeID),
		PropertyType:        result.PType,
		PropertyCategoryID:  int32(result.PropertyCategoryID),
		PropertyCategory:    result.PCategory,
		PropertyUsageID:     int32(result.PropertyUsageID),
		PropertyUsage:       result.PUsage,
		Bedrooms:            int32(result.Bedrooms),
		Bathrooms:           int32(result.Bathrooms),
		Title:               result.Title,
		Currency:            result.Currency,
		Available:           result.Available,
		Price:               result.Price,
		Deposit:             result.Deposit,
		Promoted:            result.Promoted,
		PostedDate:          timestamppb.New(result.PostedDate),
		PetsAllowed:         result.PetsAllowed,
		FreeWifi:            result.FreeWifi,
		WaterIncluded:       result.WaterIncluded,
		ElectricityIncluded: result.ElectricityIncluded,
		Latitude:            result.Latitude,
		Longitude:           result.Longitude,
	}
	// Handle nullable fields
	if result.Surburb != nil {
		property.Surburb = *result.Surburb
	}
	if result.Town != nil {
		property.Town = *result.Town
	}
	if result.PDescription != nil {
		property.Description = *result.PDescription
	}
	if result.SharingPrice != nil {
		property.SharingPrice = *result.SharingPrice
	}

	for _, row := range photos {
		property.Photos = append(property.Photos, &pb.Photo{
			Id:         row.ID.String(),
			Url:        row.PUrl,
			PropertyID: row.PropertyID.String(),
		})
	}
	db_collections.GetRedisClient().GeoAdd(propertiesGeo, &redis.GeoLocation{
		Longitude: float64(property.Longitude),
		Latitude:  float64(property.Latitude),
		Name:      property.Id,
	})
	return &property, nil
}

func (repo *PropertyRepository) GetMinimalProperty(ctx context.Context, id uuid.UUID) (*pb.MinimalProperty, error) {
	// Check for cache hit
	cachedProp, err := redis_helper.GetCachedMinimalProperty(id.String())
	if err == nil {
		return cachedProp, nil
	}
	result, err := repo.query.GetMinimalProperty(ctx, id)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			// https://www.postgresql.org/docs/9.3/errcodes-appendix.html
			switch err.Code {
			case "02000": // no_data
				return nil, status.Error(codes.NotFound, "Property not found")
			}
		}
		log.Println(err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	property := pb.MinimalProperty{
		Id:                 result.ID.String(),
		UserID:             result.UserID,
		PropertyTypeID:     int32(result.PropertyTypeID),
		PropertyType:       result.PType,
		PropertyCategoryID: int32(result.PropertyCategoryID),
		PropertyCategory:   result.PCategory,
		PropertyUsageID:    int32(result.PropertyUsageID),
		PropertyUsage:      result.PUsage,
		Bedrooms:           int32(result.Bedrooms),
		Title:              result.Title,
		Currency:           result.Currency,
		Price:              result.Price,
		Promoted:           result.Promoted,
		PostedDate:         timestamppb.New(result.PostedDate),
		Latitude:           result.Latitude,
		Longitude:          result.Longitude,
	}
	// Cache property for this location
	db_collections.GetRedisClient().GeoAdd(minimalPropertiesGeo, &redis.GeoLocation{
		Longitude: float64(property.Longitude),
		Latitude:  float64(property.Latitude),
		Name:      property.Id,
	})
	return &property, nil
}

func (repo *PropertyRepository) GetMinimalProperties(ctx context.Context, arg db.GetMinimalPropertiesParams) ([]db.GetMinimalPropertiesRow, error) {
	result, err := repo.query.GetMinimalProperties(ctx, arg)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			// https://www.postgresql.org/docs/9.3/errcodes-appendix.html
			switch err.Code {
			case "02000": // no_data
				return nil, status.Error(codes.NotFound, "No properties found")
			}
		}
		log.Println(err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return result, nil
}

func (repo *PropertyRepository) GetPromotedProperties(ctx context.Context, arg db.GetPromotedPropertiesParams) ([]db.GetPromotedPropertiesRow, error) {
	result, err := repo.query.GetPromotedProperties(ctx, arg)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			// https://www.postgresql.org/docs/9.3/errcodes-appendix.html
			switch err.Code {
			case "02000": // no_data
				return nil, status.Error(codes.NotFound, "No properties found")
			}
		}
		log.Println(err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return result, nil
}

func (repo *PropertyRepository) GetUserProperties(ctx context.Context, id string) ([]db.GetUserPropertiesRow, error) {
	result, err := repo.query.GetUserProperties(ctx, id)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			// https://www.postgresql.org/docs/9.3/errcodes-appendix.html
			switch err.Code {
			case "02000": // no_data
				return nil, status.Error(codes.NotFound, "No properties found")
			}
		}
		log.Println(err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return result, nil
}

func (repo *PropertyRepository) GetIdsOfPropertiesByUser(ctx context.Context, id string) ([]uuid.UUID, error) {
	result, err := repo.query.GetIdsOfPropertiesByUser(ctx, id)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			// https://www.postgresql.org/docs/9.3/errcodes-appendix.html
			switch err.Code {
			case "02000": // no_data
				return nil, status.Error(codes.NotFound, "No properties found")
			}
		}
		log.Println(err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return result, nil
}

func (repo *PropertyRepository) GetFeaturedAreas(ctx context.Context, countryCode string) ([]db.GetFeaturedAreasRow, error) {
	result, err := repo.query.GetFeaturedAreas(ctx, countryCode)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			// https://www.postgresql.org/docs/9.3/errcodes-appendix.html
			switch err.Code {
			case "02000": // no_data
				return nil, status.Error(codes.NotFound, "No featured areas found")
			}
		}
		log.Println(err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return result, nil
}

func (repo *PropertyRepository) GetPropertiesWithinRadius(ctx context.Context, arg db.GetPropertiesWithinRadiusParams) ([]*pb.Property, error) {
	// check cache
	cacheResults, err := db_collections.GetRedisClient().GeoRadius(propertiesGeo, arg.LlToEarth_2, arg.LlToEarth, &redis.GeoRadiusQuery{
		Radius: arg.EarthBox,
		Unit:   "m",
		Sort:   "ASC",
	}).Result()

	var properties []*pb.Property
	if err == nil && len(cacheResults) != 0 {
		// Cache hit
		for _, geo := range cacheResults {
			prop, err := redis_helper.GetCachedSingleProperty(geo.Name)
			if err != nil {
				// Cache miss
				id, err := uuid.Parse(geo.Name)
				if err != nil {
					log.Println(err)
					continue
				}
				// fetch & cache
				prop, err = repo.GetProperty(ctx, id)
				if err != nil {
					continue
				}
			}
			properties = append(properties, prop)
		}
		return properties, nil
	}

	results, err := repo.query.GetPropertiesWithinRadius(ctx, arg)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			// https://www.postgresql.org/docs/9.3/errcodes-appendix.html
			switch err.Code {
			case "02000": // no_data
				return nil, status.Error(codes.NotFound, "No properties found")
			}
		}
		log.Println(err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	for _, result := range results {
		photos, err := repo.GetPropertyPhotos(ctx, result.ID)
		if err != nil {
			return nil, err
		}
		property := pb.Property{
			Id:                  result.ID.String(),
			UserID:              result.UserID,
			PropertyTypeID:      int32(result.PropertyTypeID),
			PropertyType:        result.PType,
			PropertyCategoryID:  int32(result.PropertyCategoryID),
			PropertyCategory:    result.PCategory,
			PropertyUsageID:     int32(result.PropertyUsageID),
			PropertyUsage:       result.PUsage,
			Bedrooms:            int32(result.Bedrooms),
			Bathrooms:           int32(result.Bathrooms),
			Title:               result.Title,
			Currency:            result.Currency,
			Available:           result.Available,
			Price:               result.Price,
			Deposit:             result.Deposit,
			Promoted:            result.Promoted,
			PostedDate:          timestamppb.New(result.PostedDate),
			PetsAllowed:         result.PetsAllowed,
			FreeWifi:            result.FreeWifi,
			WaterIncluded:       result.WaterIncluded,
			ElectricityIncluded: result.ElectricityIncluded,
			Latitude:            result.Latitude,
			Longitude:           result.Longitude,
		}
		// Handle nullable fields
		if result.Surburb != nil {
			property.Surburb = *result.Surburb
		}
		if result.Town != nil {
			property.Town = *result.Town
		}
		if result.PDescription != nil {
			property.Description = *result.PDescription
		}
		if result.SharingPrice != nil {
			property.SharingPrice = *result.SharingPrice
		}
		for _, row := range photos {
			property.Photos = append(property.Photos, &pb.Photo{
				Id:         row.ID.String(),
				Url:        row.PUrl,
				PropertyID: row.PropertyID.String(),
			})
		}
		properties = append(properties, &property)
	}
	return properties, nil
}

func (repo *PropertyRepository) GetMinimalInfoPropertiesWithinRadius(ctx context.Context, arg db.GetMinimalInfoPropertiesWithinRadiusParams, promotedOnly bool) ([]*pb.MinimalProperty, error) {
	// check cache
	cacheResults, err := db_collections.GetRedisClient().GeoRadius(propertiesGeo, arg.LlToEarth_2, arg.LlToEarth, &redis.GeoRadiusQuery{
		Radius: arg.EarthBox,
		Unit:   "m",
		Sort:   "ASC",
	}).Result()
	var properties []*pb.MinimalProperty
	if err == nil && len(cacheResults) != 0 {
		// Cache hit
		for _, geo := range cacheResults {
			prop, err := redis_helper.GetCachedMinimalProperty(geo.Name)
			if err != nil {
				// Cache miss
				id, err := uuid.Parse(geo.Name)
				if err != nil {
					log.Println(err)
					continue
				}
				// fetch & cache
				prop, err = repo.GetMinimalProperty(ctx, id)
				if err != nil {
					continue
				}
			}
			if promotedOnly {
				// Only add those promoted
				if prop.Promoted {
					properties = append(properties, prop)
				}
			} else {
				properties = append(properties, prop)
			}
		}
		return properties, nil
	}

	results, err := repo.query.GetMinimalInfoPropertiesWithinRadius(ctx, arg)

	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			// https://www.postgresql.org/docs/9.3/errcodes-appendix.html
			switch err.Code {
			case "02000": // no_data
				return nil, status.Error(codes.NotFound, "No properties found")
			}
		}
		log.Println(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	for _, result := range results {
		if promotedOnly {
			// Only add those promoted
			if !result.Promoted {
				continue
			}
		}
		photos, err := repo.GetPropertyPhotos(ctx, result.ID)
		if err != nil {
			return nil, err
		}
		property := pb.MinimalProperty{
			Id:                 result.ID.String(),
			UserID:             result.UserID,
			PropertyTypeID:     int32(result.PropertyTypeID),
			PropertyType:       result.PType,
			PropertyCategoryID: int32(result.PropertyCategoryID),
			PropertyCategory:   result.PCategory,
			PropertyUsageID:    int32(result.PropertyUsageID),
			PropertyUsage:      result.PUsage,
			Bedrooms:           int32(result.Bedrooms),
			Title:              result.Title,
			Currency:           result.Currency,
			Price:              result.Price,
			Promoted:           result.Promoted,
			PostedDate:         timestamppb.New(result.PostedDate),
			Latitude:           result.Latitude,
			Longitude:          result.Longitude,
		}
		// Handle nullable fields
		if result.Town != nil {
			property.Town = *result.Town
		}
		for _, row := range photos {
			property.Photos = append(property.Photos, &pb.Photo{
				Id:         row.ID.String(),
				Url:        row.PUrl,
				PropertyID: row.PropertyID.String(),
			})
		}
		properties = append(properties, &property)
	}

	return properties, nil
}

func (repo *PropertyRepository) GetPropertyPhotos(ctx context.Context, id uuid.UUID) ([]db.GetPropertyPhotosRow, error) {
	result, err := repo.query.GetPropertyPhotos(ctx, id)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			// https://www.postgresql.org/docs/9.3/errcodes-appendix.html
			switch err.Code {
			case "02000": // no_data
				return nil, status.Error(codes.NotFound, "No photos found")
			}
		}
		log.Println(err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return result, nil
}

func (repo *PropertyRepository) DeleteProperty(ctx context.Context, id uuid.UUID) error {
	err := repo.query.DeleteProperty(ctx, id)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			// https://www.postgresql.org/docs/9.3/errcodes-appendix.html
			switch err.Code {
			case "02000": // no_data
				return status.Error(codes.NotFound, "Property not found")
			}
		}
		log.Println(err)
		return status.Error(codes.Internal, "Could not delete property")
	}
	redis_helper.FlushProperty(id.String())

	return err

}
