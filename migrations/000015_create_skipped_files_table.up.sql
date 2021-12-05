CREATE TABLE sec.skipped_files (
    id serial PRIMARY KEY,
    url text,
    created_at timestamp with time zone DEFAULT NOW()
);