// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.14.0
// source: property.sql

package db

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const createProperty = `-- name: CreateProperty :one
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
RETURNING id, user_id, property_type_id, property_category_id, property_usage_id, bedrooms, bathrooms, surburb, town, title, p_description, currency, available, price, deposit, sharing_price, promoted, posted_date, pets_allowed, free_wifi, water_included, electricity_included, latitude, longitude
`

type CreatePropertyParams struct {
	UserID              string   `json:"user_id"`
	PropertyTypeID      int16    `json:"property_type_id"`
	PropertyCategoryID  int16    `json:"property_category_id"`
	PropertyUsageID     int16    `json:"property_usage_id"`
	Bedrooms            int16    `json:"bedrooms"`
	Bathrooms           int16    `json:"bathrooms"`
	Surburb             *string  `json:"surburb"`
	Town                *string  `json:"town"`
	Title               string   `json:"title"`
	PDescription        *string  `json:"p_description"`
	Currency            string   `json:"currency"`
	Available           bool     `json:"available"`
	Price               float32  `json:"price"`
	Deposit             float32  `json:"deposit"`
	SharingPrice        *float32 `json:"sharing_price"`
	PetsAllowed         bool     `json:"pets_allowed"`
	FreeWifi            bool     `json:"free_wifi"`
	WaterIncluded       bool     `json:"water_included"`
	ElectricityIncluded bool     `json:"electricity_included"`
	Latitude            float32  `json:"latitude"`
	Longitude           float32  `json:"longitude"`
}

func (q *Queries) CreateProperty(ctx context.Context, arg CreatePropertyParams) (Property, error) {
	row := q.queryRow(ctx, q.createPropertyStmt, createProperty,
		arg.UserID,
		arg.PropertyTypeID,
		arg.PropertyCategoryID,
		arg.PropertyUsageID,
		arg.Bedrooms,
		arg.Bathrooms,
		arg.Surburb,
		arg.Town,
		arg.Title,
		arg.PDescription,
		arg.Currency,
		arg.Available,
		arg.Price,
		arg.Deposit,
		arg.SharingPrice,
		arg.PetsAllowed,
		arg.FreeWifi,
		arg.WaterIncluded,
		arg.ElectricityIncluded,
		arg.Latitude,
		arg.Longitude,
	)
	var i Property
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.PropertyTypeID,
		&i.PropertyCategoryID,
		&i.PropertyUsageID,
		&i.Bedrooms,
		&i.Bathrooms,
		&i.Surburb,
		&i.Town,
		&i.Title,
		&i.PDescription,
		&i.Currency,
		&i.Available,
		&i.Price,
		&i.Deposit,
		&i.SharingPrice,
		&i.Promoted,
		&i.PostedDate,
		&i.PetsAllowed,
		&i.FreeWifi,
		&i.WaterIncluded,
		&i.ElectricityIncluded,
		&i.Latitude,
		&i.Longitude,
	)
	return i, err
}

const createPropertyPhoto = `-- name: CreatePropertyPhoto :one
INSERT INTO property_photos (property_id, p_url)
VALUES ($1, $2)
RETURNING id, property_id, p_url
`

type CreatePropertyPhotoParams struct {
	PropertyID uuid.UUID `json:"property_id"`
	PUrl       string    `json:"p_url"`
}

func (q *Queries) CreatePropertyPhoto(ctx context.Context, arg CreatePropertyPhotoParams) (PropertyPhoto, error) {
	row := q.queryRow(ctx, q.createPropertyPhotoStmt, createPropertyPhoto, arg.PropertyID, arg.PUrl)
	var i PropertyPhoto
	err := row.Scan(&i.ID, &i.PropertyID, &i.PUrl)
	return i, err
}

const deleteProperty = `-- name: DeleteProperty :exec
DELETE FROM properties WHERE id = $1
`

func (q *Queries) DeleteProperty(ctx context.Context, id uuid.UUID) error {
	_, err := q.exec(ctx, q.deletePropertyStmt, deleteProperty, id)
	return err
}

const getFeaturedAreas = `-- name: GetFeaturedAreas :many
SELECT
	title,
	photo_url,
	latitude,
	longitude
FROM
	featured_areas
WHERE
	country = $1
`

type GetFeaturedAreasRow struct {
	Title     string  `json:"title"`
	PhotoUrl  string  `json:"photo_url"`
	Latitude  float32 `json:"latitude"`
	Longitude float32 `json:"longitude"`
}

func (q *Queries) GetFeaturedAreas(ctx context.Context, country string) ([]GetFeaturedAreasRow, error) {
	rows, err := q.query(ctx, q.getFeaturedAreasStmt, getFeaturedAreas, country)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetFeaturedAreasRow
	for rows.Next() {
		var i GetFeaturedAreasRow
		if err := rows.Scan(
			&i.Title,
			&i.PhotoUrl,
			&i.Latitude,
			&i.Longitude,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getIdsOfPropertiesByUser = `-- name: GetIdsOfPropertiesByUser :many
SELECT
	p.id
FROM
	properties p
WHERE
	p.user_id = $1
`

func (q *Queries) GetIdsOfPropertiesByUser(ctx context.Context, userID string) ([]uuid.UUID, error) {
	rows, err := q.query(ctx, q.getIdsOfPropertiesByUserStmt, getIdsOfPropertiesByUser, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		items = append(items, id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getMinimalInfoPropertiesWithinRadius = `-- name: GetMinimalInfoPropertiesWithinRadius :many
SELECT
	p.id,
	ptype.p_type,
	p.property_type_id,
	pcat.p_category,
	p.property_category_id,
	pusage.p_usage,
	p.property_usage_id,
	p.bedrooms,
	p.title,
	p.currency,
	p.promoted,
	p.price,
	p.posted_date,
	p.user_id,
	p.latitude,
	p.longitude,
	p.town
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
`

type GetMinimalInfoPropertiesWithinRadiusParams struct {
	LlToEarth   float64 `json:"ll_to_earth"`
	LlToEarth_2 float64 `json:"ll_to_earth_2"`
	EarthBox    float64 `json:"earth_box"`
}

type GetMinimalInfoPropertiesWithinRadiusRow struct {
	ID                 uuid.UUID `json:"id"`
	PType              string    `json:"p_type"`
	PropertyTypeID     int16     `json:"property_type_id"`
	PCategory          string    `json:"p_category"`
	PropertyCategoryID int16     `json:"property_category_id"`
	PUsage             string    `json:"p_usage"`
	PropertyUsageID    int16     `json:"property_usage_id"`
	Bedrooms           int16     `json:"bedrooms"`
	Title              string    `json:"title"`
	Currency           string    `json:"currency"`
	Promoted           bool      `json:"promoted"`
	Price              float32   `json:"price"`
	PostedDate         time.Time `json:"posted_date"`
	UserID             string    `json:"user_id"`
	Latitude           float32   `json:"latitude"`
	Longitude          float32   `json:"longitude"`
	Town               *string   `json:"town"`
}

func (q *Queries) GetMinimalInfoPropertiesWithinRadius(ctx context.Context, arg GetMinimalInfoPropertiesWithinRadiusParams) ([]GetMinimalInfoPropertiesWithinRadiusRow, error) {
	rows, err := q.query(ctx, q.getMinimalInfoPropertiesWithinRadiusStmt, getMinimalInfoPropertiesWithinRadius, arg.LlToEarth, arg.LlToEarth_2, arg.EarthBox)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetMinimalInfoPropertiesWithinRadiusRow
	for rows.Next() {
		var i GetMinimalInfoPropertiesWithinRadiusRow
		if err := rows.Scan(
			&i.ID,
			&i.PType,
			&i.PropertyTypeID,
			&i.PCategory,
			&i.PropertyCategoryID,
			&i.PUsage,
			&i.PropertyUsageID,
			&i.Bedrooms,
			&i.Title,
			&i.Currency,
			&i.Promoted,
			&i.Price,
			&i.PostedDate,
			&i.UserID,
			&i.Latitude,
			&i.Longitude,
			&i.Town,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getMinimalProperties = `-- name: GetMinimalProperties :many
SELECT
	p.id,
	ptype.p_type,
	p.property_type_id,
	pcat.p_category,
	p.property_category_id,
	pusage.p_usage,
	p.property_usage_id,
	p.bedrooms,
	p.title,
	p.currency,
	p.price,
	p.posted_date,
	p.user_id,
	p.town
FROM
	properties p,
	lateral(SELECT id, p_type FROM property_type WHERE id = p.property_type_id) as ptype,
	lateral(SELECT id, p_category FROM property_category WHERE id = p.property_category_id) as pcat,
	lateral(SELECT id, p_usage FROM property_usage WHERE id = p.property_usage_id) as pusage
WHERE
	earth_box(ll_to_earth($1, $2), $3) @> ll_to_earth(latitude, longitude)
`

type GetMinimalPropertiesParams struct {
	LlToEarth   float64 `json:"ll_to_earth"`
	LlToEarth_2 float64 `json:"ll_to_earth_2"`
	EarthBox    float64 `json:"earth_box"`
}

type GetMinimalPropertiesRow struct {
	ID                 uuid.UUID `json:"id"`
	PType              string    `json:"p_type"`
	PropertyTypeID     int16     `json:"property_type_id"`
	PCategory          string    `json:"p_category"`
	PropertyCategoryID int16     `json:"property_category_id"`
	PUsage             string    `json:"p_usage"`
	PropertyUsageID    int16     `json:"property_usage_id"`
	Bedrooms           int16     `json:"bedrooms"`
	Title              string    `json:"title"`
	Currency           string    `json:"currency"`
	Price              float32   `json:"price"`
	PostedDate         time.Time `json:"posted_date"`
	UserID             string    `json:"user_id"`
	Town               *string   `json:"town"`
}

func (q *Queries) GetMinimalProperties(ctx context.Context, arg GetMinimalPropertiesParams) ([]GetMinimalPropertiesRow, error) {
	rows, err := q.query(ctx, q.getMinimalPropertiesStmt, getMinimalProperties, arg.LlToEarth, arg.LlToEarth_2, arg.EarthBox)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetMinimalPropertiesRow
	for rows.Next() {
		var i GetMinimalPropertiesRow
		if err := rows.Scan(
			&i.ID,
			&i.PType,
			&i.PropertyTypeID,
			&i.PCategory,
			&i.PropertyCategoryID,
			&i.PUsage,
			&i.PropertyUsageID,
			&i.Bedrooms,
			&i.Title,
			&i.Currency,
			&i.Price,
			&i.PostedDate,
			&i.UserID,
			&i.Town,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getMinimalProperty = `-- name: GetMinimalProperty :one
SELECT
	p.id,
	ptype.p_type,
	p.property_type_id,
	pcat.p_category,
	p.property_category_id,
	pusage.p_usage,
	p.property_usage_id,
	p.bedrooms,
	p.title,
	p.currency,
	p.price,
	p.promoted,
	p.posted_date,
	p.user_id,
	p.latitude,
	p.longitude,
	p.town
FROM
	properties p,
	lateral(SELECT id, p_type FROM property_type WHERE id = p.property_type_id) as ptype,
	lateral(SELECT id, p_category FROM property_category WHERE id = p.property_category_id) as pcat,
	lateral(SELECT id, p_usage FROM property_usage WHERE id = p.property_usage_id) as pusage

WHERE
	p.id = $1
`

type GetMinimalPropertyRow struct {
	ID                 uuid.UUID `json:"id"`
	PType              string    `json:"p_type"`
	PropertyTypeID     int16     `json:"property_type_id"`
	PCategory          string    `json:"p_category"`
	PropertyCategoryID int16     `json:"property_category_id"`
	PUsage             string    `json:"p_usage"`
	PropertyUsageID    int16     `json:"property_usage_id"`
	Bedrooms           int16     `json:"bedrooms"`
	Title              string    `json:"title"`
	Currency           string    `json:"currency"`
	Price              float32   `json:"price"`
	Promoted           bool      `json:"promoted"`
	PostedDate         time.Time `json:"posted_date"`
	UserID             string    `json:"user_id"`
	Latitude           float32   `json:"latitude"`
	Longitude          float32   `json:"longitude"`
	Town               *string   `json:"town"`
}

func (q *Queries) GetMinimalProperty(ctx context.Context, id uuid.UUID) (GetMinimalPropertyRow, error) {
	row := q.queryRow(ctx, q.getMinimalPropertyStmt, getMinimalProperty, id)
	var i GetMinimalPropertyRow
	err := row.Scan(
		&i.ID,
		&i.PType,
		&i.PropertyTypeID,
		&i.PCategory,
		&i.PropertyCategoryID,
		&i.PUsage,
		&i.PropertyUsageID,
		&i.Bedrooms,
		&i.Title,
		&i.Currency,
		&i.Price,
		&i.Promoted,
		&i.PostedDate,
		&i.UserID,
		&i.Latitude,
		&i.Longitude,
		&i.Town,
	)
	return i, err
}

const getPromotedProperties = `-- name: GetPromotedProperties :many
SELECT
	p.id,
	ptype.p_type,
	p.property_type_id,
	pcat.p_category,
	p.property_category_id,
	pusage.p_usage,
	p.property_usage_id,
	p.bedrooms,
	p.title,
	p.currency,
	p.price,
	p.posted_date,
	p.user_id,
	p.town
FROM
	properties p,
	lateral(SELECT id, p_type FROM property_type WHERE id = p.property_type_id) as ptype,
	lateral(SELECT id, p_category FROM property_category WHERE id = p.property_category_id) as pcat,
	lateral(SELECT id, p_usage FROM property_usage WHERE id = p.property_usage_id) as pusage
WHERE
	p.promoted = true AND
	earth_box(ll_to_earth($1, $2), $3) @> ll_to_earth(latitude, longitude)
`

type GetPromotedPropertiesParams struct {
	LlToEarth   float64 `json:"ll_to_earth"`
	LlToEarth_2 float64 `json:"ll_to_earth_2"`
	EarthBox    float64 `json:"earth_box"`
}

type GetPromotedPropertiesRow struct {
	ID                 uuid.UUID `json:"id"`
	PType              string    `json:"p_type"`
	PropertyTypeID     int16     `json:"property_type_id"`
	PCategory          string    `json:"p_category"`
	PropertyCategoryID int16     `json:"property_category_id"`
	PUsage             string    `json:"p_usage"`
	PropertyUsageID    int16     `json:"property_usage_id"`
	Bedrooms           int16     `json:"bedrooms"`
	Title              string    `json:"title"`
	Currency           string    `json:"currency"`
	Price              float32   `json:"price"`
	PostedDate         time.Time `json:"posted_date"`
	UserID             string    `json:"user_id"`
	Town               *string   `json:"town"`
}

func (q *Queries) GetPromotedProperties(ctx context.Context, arg GetPromotedPropertiesParams) ([]GetPromotedPropertiesRow, error) {
	rows, err := q.query(ctx, q.getPromotedPropertiesStmt, getPromotedProperties, arg.LlToEarth, arg.LlToEarth_2, arg.EarthBox)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetPromotedPropertiesRow
	for rows.Next() {
		var i GetPromotedPropertiesRow
		if err := rows.Scan(
			&i.ID,
			&i.PType,
			&i.PropertyTypeID,
			&i.PCategory,
			&i.PropertyCategoryID,
			&i.PUsage,
			&i.PropertyUsageID,
			&i.Bedrooms,
			&i.Title,
			&i.Currency,
			&i.Price,
			&i.PostedDate,
			&i.UserID,
			&i.Town,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getPropertiesWithinRadius = `-- name: GetPropertiesWithinRadius :many
SELECT
	p.id,
	p.user_id,
	ptype.p_type,
	p.property_type_id,
	pcat.p_category,
	p.property_category_id,
	pusage.p_usage,
	p.property_usage_id,
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
	p.promoted,
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
`

type GetPropertiesWithinRadiusParams struct {
	LlToEarth   float64 `json:"ll_to_earth"`
	LlToEarth_2 float64 `json:"ll_to_earth_2"`
	EarthBox    float64 `json:"earth_box"`
}

type GetPropertiesWithinRadiusRow struct {
	ID                  uuid.UUID `json:"id"`
	UserID              string    `json:"user_id"`
	PType               string    `json:"p_type"`
	PropertyTypeID      int16     `json:"property_type_id"`
	PCategory           string    `json:"p_category"`
	PropertyCategoryID  int16     `json:"property_category_id"`
	PUsage              string    `json:"p_usage"`
	PropertyUsageID     int16     `json:"property_usage_id"`
	Bedrooms            int16     `json:"bedrooms"`
	Bathrooms           int16     `json:"bathrooms"`
	Surburb             *string   `json:"surburb"`
	Town                *string   `json:"town"`
	Title               string    `json:"title"`
	PDescription        *string   `json:"p_description"`
	Currency            string    `json:"currency"`
	Available           bool      `json:"available"`
	Price               float32   `json:"price"`
	Deposit             float32   `json:"deposit"`
	Promoted            bool      `json:"promoted"`
	SharingPrice        *float32  `json:"sharing_price"`
	PetsAllowed         bool      `json:"pets_allowed"`
	FreeWifi            bool      `json:"free_wifi"`
	WaterIncluded       bool      `json:"water_included"`
	ElectricityIncluded bool      `json:"electricity_included"`
	Latitude            float32   `json:"latitude"`
	Longitude           float32   `json:"longitude"`
	PostedDate          time.Time `json:"posted_date"`
}

func (q *Queries) GetPropertiesWithinRadius(ctx context.Context, arg GetPropertiesWithinRadiusParams) ([]GetPropertiesWithinRadiusRow, error) {
	rows, err := q.query(ctx, q.getPropertiesWithinRadiusStmt, getPropertiesWithinRadius, arg.LlToEarth, arg.LlToEarth_2, arg.EarthBox)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetPropertiesWithinRadiusRow
	for rows.Next() {
		var i GetPropertiesWithinRadiusRow
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.PType,
			&i.PropertyTypeID,
			&i.PCategory,
			&i.PropertyCategoryID,
			&i.PUsage,
			&i.PropertyUsageID,
			&i.Bedrooms,
			&i.Bathrooms,
			&i.Surburb,
			&i.Town,
			&i.Title,
			&i.PDescription,
			&i.Currency,
			&i.Available,
			&i.Price,
			&i.Deposit,
			&i.Promoted,
			&i.SharingPrice,
			&i.PetsAllowed,
			&i.FreeWifi,
			&i.WaterIncluded,
			&i.ElectricityIncluded,
			&i.Latitude,
			&i.Longitude,
			&i.PostedDate,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getProperty = `-- name: GetProperty :one
SELECT
	p.id,
	p.user_id,
	ptype.p_type,
	p.property_type_id,
	pcat.p_category,
	p.property_category_id,
	pusage.p_usage,
	p.property_usage_id,
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
	p.promoted,
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
	p.id = $1
`

type GetPropertyRow struct {
	ID                  uuid.UUID `json:"id"`
	UserID              string    `json:"user_id"`
	PType               string    `json:"p_type"`
	PropertyTypeID      int16     `json:"property_type_id"`
	PCategory           string    `json:"p_category"`
	PropertyCategoryID  int16     `json:"property_category_id"`
	PUsage              string    `json:"p_usage"`
	PropertyUsageID     int16     `json:"property_usage_id"`
	Bedrooms            int16     `json:"bedrooms"`
	Bathrooms           int16     `json:"bathrooms"`
	Surburb             *string   `json:"surburb"`
	Town                *string   `json:"town"`
	Title               string    `json:"title"`
	PDescription        *string   `json:"p_description"`
	Currency            string    `json:"currency"`
	Available           bool      `json:"available"`
	Price               float32   `json:"price"`
	Deposit             float32   `json:"deposit"`
	SharingPrice        *float32  `json:"sharing_price"`
	PetsAllowed         bool      `json:"pets_allowed"`
	Promoted            bool      `json:"promoted"`
	FreeWifi            bool      `json:"free_wifi"`
	WaterIncluded       bool      `json:"water_included"`
	ElectricityIncluded bool      `json:"electricity_included"`
	Latitude            float32   `json:"latitude"`
	Longitude           float32   `json:"longitude"`
	PostedDate          time.Time `json:"posted_date"`
}

func (q *Queries) GetProperty(ctx context.Context, id uuid.UUID) (GetPropertyRow, error) {
	row := q.queryRow(ctx, q.getPropertyStmt, getProperty, id)
	var i GetPropertyRow
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.PType,
		&i.PropertyTypeID,
		&i.PCategory,
		&i.PropertyCategoryID,
		&i.PUsage,
		&i.PropertyUsageID,
		&i.Bedrooms,
		&i.Bathrooms,
		&i.Surburb,
		&i.Town,
		&i.Title,
		&i.PDescription,
		&i.Currency,
		&i.Available,
		&i.Price,
		&i.Deposit,
		&i.SharingPrice,
		&i.PetsAllowed,
		&i.Promoted,
		&i.FreeWifi,
		&i.WaterIncluded,
		&i.ElectricityIncluded,
		&i.Latitude,
		&i.Longitude,
		&i.PostedDate,
	)
	return i, err
}

const getPropertyPhotos = `-- name: GetPropertyPhotos :many
SELECT
	id,
	p_url,
	property_id
FROM
	property_photos
WHERE property_id = $1
`

type GetPropertyPhotosRow struct {
	ID         uuid.UUID `json:"id"`
	PUrl       string    `json:"p_url"`
	PropertyID uuid.UUID `json:"property_id"`
}

func (q *Queries) GetPropertyPhotos(ctx context.Context, propertyID uuid.UUID) ([]GetPropertyPhotosRow, error) {
	rows, err := q.query(ctx, q.getPropertyPhotosStmt, getPropertyPhotos, propertyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetPropertyPhotosRow
	for rows.Next() {
		var i GetPropertyPhotosRow
		if err := rows.Scan(&i.ID, &i.PUrl, &i.PropertyID); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getUserProperties = `-- name: GetUserProperties :many
SELECT
	p.id,
	p.user_id,
	ptype.p_type,
	p.property_type_id,
	pcat.p_category,
	p.property_category_id,
	pusage.p_usage,
	p.property_usage_id,
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
	p.user_id = $1
`

type GetUserPropertiesRow struct {
	ID                  uuid.UUID `json:"id"`
	UserID              string    `json:"user_id"`
	PType               string    `json:"p_type"`
	PropertyTypeID      int16     `json:"property_type_id"`
	PCategory           string    `json:"p_category"`
	PropertyCategoryID  int16     `json:"property_category_id"`
	PUsage              string    `json:"p_usage"`
	PropertyUsageID     int16     `json:"property_usage_id"`
	Bedrooms            int16     `json:"bedrooms"`
	Bathrooms           int16     `json:"bathrooms"`
	Surburb             *string   `json:"surburb"`
	Town                *string   `json:"town"`
	Title               string    `json:"title"`
	PDescription        *string   `json:"p_description"`
	Currency            string    `json:"currency"`
	Available           bool      `json:"available"`
	Price               float32   `json:"price"`
	Deposit             float32   `json:"deposit"`
	SharingPrice        *float32  `json:"sharing_price"`
	PetsAllowed         bool      `json:"pets_allowed"`
	FreeWifi            bool      `json:"free_wifi"`
	WaterIncluded       bool      `json:"water_included"`
	ElectricityIncluded bool      `json:"electricity_included"`
	Latitude            float32   `json:"latitude"`
	Longitude           float32   `json:"longitude"`
	PostedDate          time.Time `json:"posted_date"`
}

func (q *Queries) GetUserProperties(ctx context.Context, userID string) ([]GetUserPropertiesRow, error) {
	rows, err := q.query(ctx, q.getUserPropertiesStmt, getUserProperties, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetUserPropertiesRow
	for rows.Next() {
		var i GetUserPropertiesRow
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.PType,
			&i.PropertyTypeID,
			&i.PCategory,
			&i.PropertyCategoryID,
			&i.PUsage,
			&i.PropertyUsageID,
			&i.Bedrooms,
			&i.Bathrooms,
			&i.Surburb,
			&i.Town,
			&i.Title,
			&i.PDescription,
			&i.Currency,
			&i.Available,
			&i.Price,
			&i.Deposit,
			&i.SharingPrice,
			&i.PetsAllowed,
			&i.FreeWifi,
			&i.WaterIncluded,
			&i.ElectricityIncluded,
			&i.Latitude,
			&i.Longitude,
			&i.PostedDate,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
