CREATE TABLE IF NOT EXISTS workflow_history (
    id UUID PRIMARY KEY,
    created_at timestamp NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    type varchar(50) NOT NULL,
    input JSONB NOT NULL DEFAULT '{}',
    input_previous JSONB NOT NULL DEFAULT '{}',
    step_name varchar(50) NOT NULL,
    workflow_id UUID NOT NULL references workflow(id),
    request_id varchar(50) NULL,
    workflow_status varchar(50) NOT NULL,
    workflow_error TEXT NULL,
    workflow_error_kind varchar(50) NULL
);
