-- +goose Up
-- +goose StatementBegin
CREATE TABLE public.subscription
(
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_name VARCHAR(100) NOT NULL,
    price        INT          NOT NULL,
    start_date   DATE         NOT NULL,
    end_date     DATE,
    "user"       UUID         NOT NULL,
    UNIQUE ("user", service_name, start_date)
);
CREATE INDEX idx_subscription_user_service_start ON public.subscription ("user", service_name, start_date);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE public.subscription;
-- +goose StatementEnd
