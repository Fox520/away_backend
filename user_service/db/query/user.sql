-- name: CreateUser :one
INSERT INTO users (
	id, username, email, device_token, bio, verified, s_status, profile_picture_url
)
VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: GetFullUser :one
SELECT
	users.id,
	users.username,
	users.email,
	users.bio,
	users.device_token,
	users.verified,
	users.s_status,
	users.createdat,
	users.profile_picture_url,
	subscription_status.s_description
FROM   users, subscription_status
WHERE  id = $1
	AND
		users.s_status = subscription_status.s_status
LIMIT 1;

-- name: GetMinimalUser :one
SELECT
	id,
	username,
	bio,
	verified,
	createdAt,
	profile_picture_url
FROM users
WHERE id = $1 LIMIT 1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id=$1;