CREATE TABLE sec.tag (
    id serial PRIMARY KEY,
    tag text,
    version text,
    custom text,
    abstract text,
    datatype text,
    lord text,
    crdr text,
    tlabel text,
    doc text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    CONSTRAINT tag_version UNIQUE (tag, version)
);