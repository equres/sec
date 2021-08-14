CREATE TABLE worklist (
    id serial PRIMARY KEY,
    year integer,
    month integer,
    will_download boolean,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    CONSTRAINT year_month UNIQUE (year, month)
);