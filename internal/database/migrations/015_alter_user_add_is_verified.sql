-- Write your migrate up statements here
ALTER TABLE users
ADD COLUMN is_verified BOOLEAN NOT NULL DEFAULT false;
---- create above / drop below ----

-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
