CREATE TABLE instance_test (
    id UUID,
    key String,
    created_at DateTime,
    updated_at DateTime,
    bucket String,
    size Int64,
    content_type String,
    content_length Int64
) ENGINE=MergeTree() PRIMARY KEY id ORDER BY id;
