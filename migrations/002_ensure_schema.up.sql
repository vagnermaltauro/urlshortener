
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_name = 'urls' AND table_schema = 'public'
    ) THEN
        CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

        CREATE TABLE urls (
            id BIGINT NOT NULL,
            short_code VARCHAR(11) NOT NULL,
            original_url TEXT NOT NULL,
            clicks BIGINT DEFAULT 0,
            created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            expires_at TIMESTAMPTZ NOT NULL,
            PRIMARY KEY (id, created_at),
            UNIQUE (short_code, created_at),
            CONSTRAINT check_expires_after_created CHECK (expires_at > created_at),
            CONSTRAINT check_positive_clicks CHECK (clicks >= 0)
        ) PARTITION BY RANGE (created_at);

        CREATE INDEX idx_urls_short_code ON urls (short_code);
        CREATE INDEX idx_urls_expires_at ON urls (expires_at);
        CREATE INDEX idx_urls_created_at ON urls (created_at DESC);
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS urls_2025_01 PARTITION OF urls FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
CREATE TABLE IF NOT EXISTS urls_2025_02 PARTITION OF urls FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');
CREATE TABLE IF NOT EXISTS urls_2025_03 PARTITION OF urls FOR VALUES FROM ('2025-03-01') TO ('2025-04-01');
CREATE TABLE IF NOT EXISTS urls_2025_04 PARTITION OF urls FOR VALUES FROM ('2025-04-01') TO ('2025-05-01');
CREATE TABLE IF NOT EXISTS urls_2025_05 PARTITION OF urls FOR VALUES FROM ('2025-05-01') TO ('2025-06-01');
CREATE TABLE IF NOT EXISTS urls_2025_06 PARTITION OF urls FOR VALUES FROM ('2025-06-01') TO ('2025-07-01');
CREATE TABLE IF NOT EXISTS urls_2025_07 PARTITION OF urls FOR VALUES FROM ('2025-07-01') TO ('2025-08-01');
CREATE TABLE IF NOT EXISTS urls_2025_08 PARTITION OF urls FOR VALUES FROM ('2025-08-01') TO ('2025-09-01');
CREATE TABLE IF NOT EXISTS urls_2025_09 PARTITION OF urls FOR VALUES FROM ('2025-09-01') TO ('2025-10-01');
CREATE TABLE IF NOT EXISTS urls_2025_10 PARTITION OF urls FOR VALUES FROM ('2025-10-01') TO ('2025-11-01');
CREATE TABLE IF NOT EXISTS urls_2025_11 PARTITION OF urls FOR VALUES FROM ('2025-11-01') TO ('2025-12-01');
CREATE TABLE IF NOT EXISTS urls_2025_12 PARTITION OF urls FOR VALUES FROM ('2025-12-01') TO ('2026-01-01');
CREATE TABLE IF NOT EXISTS urls_2026_01 PARTITION OF urls FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

CREATE OR REPLACE FUNCTION create_next_partition()
RETURNS void AS $$
DECLARE
    partition_date DATE;
    partition_name TEXT;
    start_date TEXT;
    end_date TEXT;
BEGIN
    partition_date := DATE_TRUNC('month', NOW() + INTERVAL '1 month');

    partition_name := 'urls_' || TO_CHAR(partition_date, 'YYYY_MM');

    start_date := partition_date::TEXT;
    end_date := (partition_date + INTERVAL '1 month')::TEXT;

    EXECUTE format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF urls FOR VALUES FROM (%L) TO (%L)',
        partition_name,
        start_date,
        end_date
    );

    RAISE NOTICE 'Created partition % for range % to %', partition_name, start_date, end_date;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION delete_expired_urls()
RETURNS TABLE(deleted_count BIGINT) AS $$
DECLARE
    rows_deleted BIGINT;
BEGIN
    DELETE FROM urls WHERE expires_at < NOW();
    GET DIAGNOSTICS rows_deleted = ROW_COUNT;
    RETURN QUERY SELECT rows_deleted;
END;
$$ LANGUAGE plpgsql;
