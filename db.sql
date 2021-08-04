CREATE TABLE public.tickers
(
    id integer NOT NULL DEFAULT nextval('tickers_id_seq'::regclass),
    ticker text COLLATE pg_catalog."default",
    cik text COLLATE pg_catalog."default",
    title text COLLATE pg_catalog."default",
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    exchange text COLLATE pg_catalog."default",
    CONSTRAINT tickers_pkey PRIMARY KEY (id),
    CONSTRAINT "ALL_UNIQUE" UNIQUE (ticker, cik, title, exchange)
);