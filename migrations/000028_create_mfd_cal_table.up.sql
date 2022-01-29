CREATE TABLE mfd.cal (
    id serial PRIMARY KEY,
    adsh text,
    grp text,
    arc text,
    negative text,
    ptag text,
    pversion text,
    ctag text,
    cversion text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    CONSTRAINT adsh_grp UNIQUE (adsh, grp, arc)
);