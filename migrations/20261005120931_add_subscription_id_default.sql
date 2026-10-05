-- +goose Up
ALTER TABLE subscriptions 
ALTER COLUMN id 
SET DEFAULT gen_random_uuid();

-- +goose Down
ALTER TABLE subscriptions 
ALTER COLUMN id 
DROP DEFAULT;
