-- Write your migrate up statements here

ALTER TABLE properties DROP COLUMN images;

CREATE TABLE property_images (
  id UUID PRIMARY KEY DEFAULT uuidv7(),
  property_id UUID NOT NULL,
  key TEXT NOT NULL,
  status TEXT DEFAULT 'uploading' NOT NULL,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,

  CONSTRAINT chk_image_status
    CHECK (status IN ('uploading', 'active', 'failed')),

  CONSTRAINT fk_image_property
    FOREIGN KEY (property_id)
    REFERENCES properties (id)
    ON DELETE CASCADE
);
---- create above / drop below ----

-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
