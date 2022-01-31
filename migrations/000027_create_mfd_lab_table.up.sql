CREATE TABLE mfd.lab (
    id serial PRIMARY KEY,
    adsh text,
    tag text,
    version text,
    CONSTRAINT adsh_tag_version UNIQUE (adsh, tag, version),
    std text,
    terse text,
    verbose_val text,
    total text,
    negated text,
    negatedterse text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);