CREATE ROLE physical_rep WITH REPLICATION LOGIN PASSWORD '12345';

SELECT pg_create_physical_replication_slot('secondary_slot');