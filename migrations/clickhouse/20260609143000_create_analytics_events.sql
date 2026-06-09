-- +goose Up
CREATE TABLE IF NOT EXISTS analytics_events (
                                                event_time DateTime,
                                                event_type String,
                                                entity_type String,
                                                entity_id UInt64,
                                                department_id Nullable(UInt64),
    metadata String
    )
    ENGINE = MergeTree
    ORDER BY (event_time, event_type, entity_type);

-- +goose Down
DROP TABLE IF EXISTS analytics_events;