-- Table structure based on data in docs/mfd.pdf
CREATE TABLE mfd.cal (
    id serial PRIMARY KEY,
    adsh text,
    grp text,
    arc text,
    CONSTRAINT adsh_grp UNIQUE (adsh, grp, arc),
    negative text,
    ptag text,
    pversion text,
    ctag text,
    cversion text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);