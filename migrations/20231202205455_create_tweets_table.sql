-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "tweets" (
    tweet_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    text VARCHAR(255) NOT NULL,
    image varchar(255) null,
    created_at   TIMESTAMP WITH TIME ZONE    NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMP WITH TIME ZONE             DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
