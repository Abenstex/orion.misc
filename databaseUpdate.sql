-- name: Update object_custom_fields
ALTER TABLE object_custom_fields ADD COLUMN IF NOT EXISTS created_by VARCHAR(50);