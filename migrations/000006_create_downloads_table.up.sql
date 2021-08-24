CREATE TABLE sec.downloads (
    id serial PRIMARY KEY,
    url text,
    etag text,
    size integer,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    CONSTRAINT url_constraint UNIQUE (url)
);