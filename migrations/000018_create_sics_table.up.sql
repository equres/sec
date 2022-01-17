CREATE TABLE sec.sics (
    id serial PRIMARY KEY,
    sic integer,
    office text,
    title text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    CONSTRAINT sic_unique UNIQUE (sic)
)