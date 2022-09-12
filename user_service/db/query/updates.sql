-- name: SetUsername :one
update users
set username = $2
where id = $1
returning *;

-- name: SetDeviceToken :one
update users
set device_token = $2
where id = $1
returning *;

-- name: SetBio :one
update users
set bio = $2
where id = $1
returning *;

-- name: SetProfilePictureUrl :one
update users
set profile_picture_url = $2
where id = $1
returning *;