-- Table structure based on data in docs/mfd.pdf
CREATE TABLE mfd.tag (
    id serial PRIMARY KEY,
    tag text,
    version text,
    CONSTRAINT tag_version UNIQUE (tag, version),
    custom text,
    abstract text,
    datatype text,
    lord text,
    tlabel text,
    doc text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);