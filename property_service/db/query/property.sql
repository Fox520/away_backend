-- name: CreateProperty :one
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
RETURNING *;

-- name: CreatePropertyPhoto :one
INSERT INTO property_photos (property_id, p_url)
VALUES ($1, $2)
RETURNING *;

-- name: GetProperty :one
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
	p.id = $1;

-- name: GetPropertiesWithinRadius :many
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
	earth_box(ll_to_earth($1, $2), $3) @> ll_to_earth(latitude, longitude);

-- name: GetMinimalInfoPropertiesWithinRadius :many
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
	earth_box(ll_to_earth($1, $2), $3) @> ll_to_earth(latitude, longitude);



-- name: GetPropertyPhotos :many
SELECT
	id,
	p_url,
	property_id
FROM
	property_photos
WHERE property_id = $1;

-- name: GetMinimalProperty :one
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
	p.id = $1;

-- name: GetMinimalProperties :many
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
	earth_box(ll_to_earth($1, $2), $3) @> ll_to_earth(latitude, longitude);

-- name: GetPromotedProperties :many
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
	earth_box(ll_to_earth($1, $2), $3) @> ll_to_earth(latitude, longitude);


-- name: GetUserProperties :many
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
	p.user_id = $1;

-- name: GetIdsOfPropertiesByUser :many
SELECT
	p.id
FROM
	properties p
WHERE
	p.user_id = $1;

-- name: DeleteProperty :exec
DELETE FROM properties WHERE id = $1;

-- name: GetFeaturedAreas :many
SELECT
	title,
	photo_url,
	latitude,
	longitude
FROM
	featured_areas
WHERE
	country = $1;