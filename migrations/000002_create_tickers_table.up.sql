CREATE TABLE sec.tickers (
    id serial PRIMARY KEY,
    ticker text,
    cik text,
    title text,
    exchange text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    CONSTRAINT "ALL_UNIQUE" UNIQUE (cik, ticker, title)
);