DROP FUNCTION IF EXISTS delete_expired_urls();
DROP FUNCTION IF EXISTS create_next_partition();

DROP TABLE IF EXISTS urls CASCADE;

DROP EXTENSION IF EXISTS pg_stat_statements;
