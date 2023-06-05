CREATE TABLE IF NOT EXISTS workflow (
    id UUID PRIMARY KEY,
    parent_id UUID NULL references workflow(id),
    status TEXT NOT NULL,
    input JSONB NOT NULL DEFAULT '{}',
    schema_name TEXT NOT NULL DEFAULT '',
    steps JSONB,
    request_id varchar(50) NULL,
    error TEXT NULL,
    error_kind varchar(50) NULL,
    created_at timestamp NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at timestamp NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC')
);
