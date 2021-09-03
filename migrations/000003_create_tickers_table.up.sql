CREATE TABLE sec.tickers (
    id serial PRIMARY KEY,
    ticker text,
    cik integer,
    title text,
    exchange text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    CONSTRAINT "ALL_UNIQUE" UNIQUE (cik, ticker, title),
    CONSTRAINT fk_cik FOREIGN KEY (cik) REFERENCES sec.ciks (cik)
);