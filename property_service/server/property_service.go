package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	auth "github.com/Fox520/away_backend/auth"
	config "github.com/Fox520/away_backend/config"
	pb "github.com/Fox520/away_backend/property_service/github.com/Fox520/away_backend/property_service/pb"
	_ "github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PropertyServiceServer struct {
	pb.UnimplementedPropertyServiceServer
	DB *sql.DB
}

func NewPropertyServiceServer(cfg config.Config) *PropertyServiceServer {
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
	fmt.Println("Successfully connected to DB!")

	return &PropertyServiceServer{
		DB: db,
	}

}

// Fetch property from db
func (server *PropertyServiceServer) GetSingleProperty(ctx context.Context, pr *pb.GetSinglePropertyRequest) (*pb.GetSinglePropertyResponse, error) {

	var propertyResponse pb.GetSinglePropertyResponse
	// Fixes: invalid memory address or nil pointer dereference
	propertyResponse.Property = &pb.Property{}

	var tempTime time.Time
	err := server.DB.QueryRow(`
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
			p.id = $1`, pr.Id).Scan(
		&propertyResponse.Property.Id,
		&propertyResponse.Property.UserID,
		&propertyResponse.Property.PropertyType,
		&propertyResponse.Property.PropertyTypeID,
		&propertyResponse.Property.PropertyCategory,
		&propertyResponse.Property.PropertyCategoryID,
		&propertyResponse.Property.PropertyUsage,
		&propertyResponse.Property.PropertyUsageID,
		&propertyResponse.Property.Bedrooms,
		&propertyResponse.Property.Bathrooms,
		&propertyResponse.Property.Surburb,
		&propertyResponse.Property.Town,
		&propertyResponse.Property.Title,
		&propertyResponse.Property.Description,
		&propertyResponse.Property.Currency,
		&propertyResponse.Property.Available,
		&propertyResponse.Property.Price,
		&propertyResponse.Property.Deposit,
		&propertyResponse.Property.SharingPrice,
		&propertyResponse.Property.PetsAllowed,
		&propertyResponse.Property.FreeWifi,
		&propertyResponse.Property.WaterIncluded,
		&propertyResponse.Property.ElectricityIncluded,
		&propertyResponse.Property.Latitude,
		&propertyResponse.Property.Longitude,
		&tempTime,
	)

	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	propertyResponse.Property.PostedDate = timestamppb.New(tempTime)
	// Add photos to response
	rows, err := server.DB.Query(`SELECT id, p_url, property_id FROM property_photos WHERE property_id = $1`, pr.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	for rows.Next() {
		id := ""
		url := ""
		propId := ""
		err = rows.Scan(&id, &url, &propId)
		if err != nil {
			continue
		}

		propertyResponse.Property.Photos = append(propertyResponse.Property.Photos, &pb.Photo{Id: id, Url: url, PropertyID: propId})

	}
	return &propertyResponse, nil
}

func (server *PropertyServiceServer) GetMultipleProperties(ctx context.Context, pm *pb.GetMultiplePropertyRequest) (*pb.GetMultiplePropertyResponse, error) {
	radius := pm.Radius
	if radius < 1 {
		radius = 5
	}
	// Convert radius (km) to meters
	radius = radius * 1000

	var propertiesResponse pb.GetMultiplePropertyResponse
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
			p.description,
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
			p.longitude
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
		pm.Latitude,
		pm.Longitude,
		radius,
	)

	if err != nil {
		return nil, err
	}
	for propertyRows.Next() {
		var property pb.Property
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
			&property.Longitude)
		if err != nil {
			continue
		}
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
		propertiesResponse.Properties = append(propertiesResponse.Properties, &property)
	}

	return &propertiesResponse, nil
}

func (server *PropertyServiceServer) GetMinimalProperties(ctx context.Context, pm *pb.GetMinimalPropertiesRequest) (*pb.GetMinimalPropertiesResponse, error) {
	radius := pm.Radius
	if radius < 1 {
		radius = 5
	}
	// Convert radius (km) to meters
	radius = radius * 1000

	var propertiesResponse pb.GetMinimalPropertiesResponse
	res, err := minimalQuery(server.DB, pm.Latitude, pm.Longitude, radius, false)
	if err != nil {
		return nil, err
	}
	propertiesResponse.Properties = res
	return &propertiesResponse, nil

}

func (server *PropertyServiceServer) GetPromotedProperties(ctx context.Context, pr *pb.PromotedRequest) (*pb.PromotedResponse, error) {
	radius := pr.Radius
	if radius < 1 {
		radius = 200
	}
	// Convert radius (km) to meters
	radius = radius * 1000

	var promotedResponse pb.PromotedResponse
	res, err := minimalQuery(server.DB, pr.Latitude, pr.Longitude, radius, true)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	promotedResponse.Properties = res

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
