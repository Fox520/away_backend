package server

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	pb "github.com/Fox520/away_backend/property_service/github.com/Fox520/away_backend/property_service/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func minimalQuery(DB *sql.DB, lat float32, lng float32, radius float32, isPromoted bool) ([]*pb.MinimalProperty, error) {
	promotedString := ""
	if isPromoted {
		promotedString = "p.promoted = true AND"
	}
	var properties []*pb.MinimalProperty
	// src https://dba.stackexchange.com/a/158422
	propertyRows, err := DB.Query(fmt.Sprintf(`
		SELECT
			p.id,
			ptype.p_type,
			ptype.id,
			pcat.p_category,
			pcat.id,
			pusage.p_usage,
			pusage.id,
			p.bedrooms,
			p.title,
			p.currency,
			p.price,
			p.posted_date,
			p.user_id
		FROM
			properties p,
			lateral(SELECT id, p_type FROM property_type WHERE id = p.property_type_id) as ptype,
			lateral(SELECT id, p_category FROM property_category WHERE id = p.property_category_id) as pcat,
			lateral(SELECT id, p_usage FROM property_usage WHERE id = p.property_usage_id) as pusage
		WHERE
			%s
			earth_box(ll_to_earth($1, $2), $3) @> ll_to_earth(latitude, longitude) 
			`, promotedString),
		lat,
		lng,
		radius,
	)

	if err != nil {
		return nil, err
	}

	for propertyRows.Next() {
		var property pb.MinimalProperty
		var tempTime time.Time
		err = propertyRows.Scan(&property.Id,
			&property.PropertyType,
			&property.PropertyTypeID,
			&property.PropertyCategory,
			&property.PropertyCategoryID,
			&property.PropertyUsage,
			&property.PropertyUsageID,
			&property.Bedrooms,
			&property.Title,
			&property.Currency,
			&property.Price,
			&tempTime,
			&property.UserID,
		)
		if err != nil {
			continue
		}
		property.PostedDate = timestamppb.New(tempTime)
		// Add photos to response
		rows, err := DB.Query(`SELECT id, p_url, property_id FROM property_photos WHERE property_id = $1`, property.Id)
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
	}
	return properties, nil
}

func topAreasQuery(DB *sql.DB, country string) ([]*pb.FeaturedArea, error) {

	var areas []*pb.FeaturedArea
	// src https://dba.stackexchange.com/a/158422
	propertyRows, err := DB.Query(`
		SELECT
			title,
			photo_url,
			latitude,
			longitude
		FROM
			top_areas
		WHERE
			country = $1`,
		strings.ToLower(country),
	)

	if err != nil {
		return nil, err
	}

	for propertyRows.Next() {
		var area pb.FeaturedArea
		err = propertyRows.Scan(&area.Title,
			&area.PhotoURL,
			&area.Latitude,
			&area.Longitude,
		)
		if err != nil {
			continue
		}
		areas = append(areas, &area)
	}
	return areas, nil
}
