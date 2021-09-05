CREATE TABLE sec.ciks (
    id serial PRIMARY KEY,
    cik integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    CONSTRAINT cik_unique UNIQUE (cik)
)