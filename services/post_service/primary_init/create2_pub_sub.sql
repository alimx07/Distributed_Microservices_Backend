CREATE ROLE logical_rep WITH REPLICATION LOGIN PASSWORD '1234';

CREATE PUBLICATION my_pub FOR TABLE posts(user_id , post_id , created_at);

GRANT SELECT ON ALL TABLES IN SCHEMA public TO logical_rep;

CREATE ROLE physical_rep WITH REPLICATION LOGIN PASSWORD '12345';

SELECT pg_create_physical_replication_slot('secondary_slot');