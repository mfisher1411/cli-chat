-- +goose Up
-- +goose StatementBegin
ALTER TABLE chat
    ADD COLUMN name       text,
    ADD COLUMN created_at timestamp NOT NULL DEFAULT now();

CREATE TABLE "chat_member"
(
    user_id   int       NOT NULL,
    chat_id   int       NOT NULL,
    joined_at timestamp NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, chat_id),
    CONSTRAINT fk_chat
        FOREIGN KEY (chat_id)
            REFERENCES chat (id)
            ON DELETE CASCADE
);

CREATE TABLE "message"
(
    id        BIGSERIAL NOT NULL PRIMARY KEY,
    sender_id int       NOT NULL,
    chat_id   int       NOT NULL,
    content   text      NOT NULL,
    sent_at   timestamp NOT NULL DEFAULT now(),
    CONSTRAINT fk_chat
        FOREIGN KEY (chat_id)
            REFERENCES chat (id)
            ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "message";
DROP TABLE IF EXISTS "chat_member";
ALTER TABLE chat
    DROP COLUMN IF EXISTS name,
    DROP COLUMN IF EXISTS created_at;
-- +goose StatementEnd
	
