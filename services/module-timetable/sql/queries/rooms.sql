-- name: CreateRoom :one
INSERT INTO timetable.rooms (name, capacity, type, features)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetRoomByID :one
SELECT * FROM timetable.rooms WHERE id = $1;

-- name: ListRooms :many
SELECT * FROM timetable.rooms WHERE is_active = true ORDER BY name LIMIT $1 OFFSET $2;

-- name: CountRooms :one
SELECT COUNT(*) FROM timetable.rooms WHERE is_active = true;
