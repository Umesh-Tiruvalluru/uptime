-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF to_regclass('public.monitor_checks') IS NOT NULL
       AND to_regclass('public.check_results') IS NULL THEN
        ALTER TABLE monitor_checks RENAME TO check_results;
    END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $$
BEGIN
    IF to_regclass('public.check_results') IS NOT NULL
       AND to_regclass('public.monitor_checks') IS NULL THEN
        ALTER TABLE check_results RENAME TO monitor_checks;
    END IF;
END $$;
-- +goose StatementEnd
