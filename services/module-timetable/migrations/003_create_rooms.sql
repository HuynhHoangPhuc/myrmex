-- +goose Up
CREATE TABLE timetable.rooms (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name      VARCHAR(100) NOT NULL UNIQUE,
    capacity  INT NOT NULL CHECK (capacity > 0),
    type      VARCHAR(50) NOT NULL DEFAULT 'classroom',
    features  TEXT[] NOT NULL DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT true
);

-- +goose Down
DROP TABLE IF EXISTS timetable.rooms;
