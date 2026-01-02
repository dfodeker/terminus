-- +goose Up

ALTER TABLE stores
  ADD COLUMN IF NOT EXISTS address TEXT,
  ADD COLUMN IF NOT EXISTS status VARCHAR(50),
  ADD COLUMN IF NOT EXISTS default_currency VARCHAR(10),
  ADD COLUMN IF NOT EXISTS timezone TEXT,
  ADD COLUMN IF NOT EXISTS plan TEXT,
  ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE,
  ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE;

ALTER TABLE stores
  ALTER COLUMN address SET DEFAULT '',
  ALTER COLUMN address SET NOT NULL,

  ALTER COLUMN status SET DEFAULT 'active',
  ALTER COLUMN status SET NOT NULL,

  ALTER COLUMN default_currency SET DEFAULT 'USD',
  ALTER COLUMN default_currency SET NOT NULL,

  ALTER COLUMN timezone SET DEFAULT 'UTC',
  ALTER COLUMN timezone SET NOT NULL,

  ALTER COLUMN created_at SET DEFAULT CURRENT_TIMESTAMP,
  ALTER COLUMN created_at SET NOT NULL,
  
  ALTER COLUMN plan SET DEFAULT 'free',
  ALTER COLUMN plan SET NOT NULL,

  ALTER COLUMN updated_at SET DEFAULT CURRENT_TIMESTAMP,
  ALTER COLUMN updated_at SET NOT NULL;

  


-- +goose Down
ALTER TABLE stores
  -- address was nullable with no default in 002
  ALTER COLUMN address DROP NOT NULL,
  ALTER COLUMN address DROP DEFAULT,

  -- status had a default but was nullable in 002
  ALTER COLUMN status DROP NOT NULL,
  ALTER COLUMN status SET DEFAULT 'active',

  -- default_currency had a default but was nullable in 002
  ALTER COLUMN default_currency DROP NOT NULL,
  ALTER COLUMN default_currency SET DEFAULT 'USD',

  -- timezone was nullable with no default in 002
  ALTER COLUMN timezone DROP NOT NULL,
  ALTER COLUMN timezone DROP DEFAULT,

  -- timestamps had defaults but were nullable in 002
  ALTER COLUMN created_at DROP NOT NULL,
  ALTER COLUMN created_at SET DEFAULT CURRENT_TIMESTAMP,

  ALTER COLUMN updated_at DROP NOT NULL,
  ALTER COLUMN updated_at SET DEFAULT CURRENT_TIMESTAMP;
