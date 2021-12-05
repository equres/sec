CREATE TABLE sec.events (
    id serial PRIMARY KEY,
    ev json,
    created_at timestamp with time zone
);