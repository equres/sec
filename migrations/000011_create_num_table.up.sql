CREATE TABLE sec.num (
    id serial PRIMARY KEY,
    adsh text,
    tag text,
    version text,
    coreg text,
    ddate text,
    qtrs text,
    uom text,
    value text,
    footnote text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    CONSTRAINT unique_keys UNIQUE (adsh, tag, version, coreg, ddate, qtrs, uom),
    FOREIGN KEY (tag, version) REFERENCES sec.tag (tag, version)
);