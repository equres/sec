CREATE TABLE public.worklist
(
    id integer NOT NULL DEFAULT nextval('worklist_id_seq'::regclass),
    year integer,
    month integer,
    will_download boolean,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    CONSTRAINT worklist_pkey PRIMARY KEY (id),
    CONSTRAINT year_month UNIQUE (year, month)
)