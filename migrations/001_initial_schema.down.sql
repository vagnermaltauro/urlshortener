-- Drop functions
DROP FUNCTION IF EXISTS delete_expired_urls();
DROP FUNCTION IF EXISTS create_next_partition();

-- Drop table (this will automatically drop all partitions)
DROP TABLE IF EXISTS urls CASCADE;

-- Drop extension
DROP EXTENSION IF EXISTS pg_stat_statements;
