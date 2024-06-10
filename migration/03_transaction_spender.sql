-- +goose Up
-- +goose StatementBegin
ALTER TABLE  "transaction" (
  ADD spender_id INT DEFAULT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
