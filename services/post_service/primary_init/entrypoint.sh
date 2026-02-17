#!/usr/bin/env bash
set -e


echo "Configuring primary for streaming replication..."

cat >> "$PGDATA/postgresql.conf" <<'EOF'
# replication settings
listen_addresses = '*'
wal_level = logical
max_wal_senders = 10
wal_keep_size = 128MB
hot_standby = on
EOF

echo "host replication physical_rep 0.0.0.0/0 md5" >> "$PGDATA/pg_hba.conf"


echo "Primary configured."