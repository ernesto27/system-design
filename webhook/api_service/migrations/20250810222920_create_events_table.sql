-- +goose Up
-- +goose StatementBegin
CREATE TABLE webhook_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id VARCHAR(255),
    source VARCHAR(255) NOT NULL,
    type VARCHAR(255) NOT NULL,
    data JSONB NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhook_events_source ON webhook_events (source);
CREATE INDEX idx_webhook_events_type ON webhook_events (type);
CREATE INDEX idx_webhook_events_timestamp ON webhook_events (timestamp);
CREATE INDEX idx_webhook_events_created_at ON webhook_events (created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS webhook_events;
-- +goose StatementEnd
