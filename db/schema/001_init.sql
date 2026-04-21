-- +goose NO TRANSACTION
-- +goose Up
PRAGMA foreign_keys = ON;
PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;

-- +goose Down
PRAGMA foreign_keys = OFF;
PRAGMA journal_mode = DELETE;
PRAGMA synchronous = FULL;
