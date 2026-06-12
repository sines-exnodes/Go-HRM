CREATE TABLE user_contracts (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id    UUID        NOT NULL REFERENCES employees(id),
    contract_type  TEXT        NOT NULL,
    signed_date    DATE        NOT NULL,
    expiry_date    DATE,
    is_endless     BOOLEAN     NOT NULL DEFAULT false,
    attachment_url TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    is_deleted     BOOLEAN     NOT NULL DEFAULT false,
    deleted_at     TIMESTAMPTZ
);

CREATE INDEX idx_user_contracts_employee_id
    ON user_contracts(employee_id)
    WHERE is_deleted = false;

CREATE TRIGGER set_updated_at
    BEFORE UPDATE ON user_contracts
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
