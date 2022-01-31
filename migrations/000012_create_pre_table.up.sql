-- Table structure based on data in docs/fsds.pdf
CREATE TABLE sec.pre (
    id serial PRIMARY KEY,
    adsh text,
    report text,
    line text,
    stmt text,
    inpth text,
    rfile text,
    tag text,
    version text,
    plabel text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    CONSTRAINT adsh_report_line UNIQUE (adsh, report, line),
    FOREIGN KEY (tag, version) REFERENCES sec.tag (tag, version)
);