CREATE TABLE reset_token (
	id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
	token TEXT UNIQUE NOT NULL,
	used boolean NOT NULL DEFAULT FALSE,
    created_at timestamptz NOT NULL DEFAULT current_timestamp,
    updated_at timestamptz NOT NULL DEFAULT current_timestamp,
    deleted_at timestamptz 
);